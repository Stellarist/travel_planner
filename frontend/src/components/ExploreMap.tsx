import { useState, useEffect, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import AMapLoader from '@amap/amap-jsapi-loader'
import configJson from '../../../config.json'
import './ExploreMap.css'
import type { Attraction } from '../shared/types'
import { ATTRACTION_TYPES } from '../shared/constants'

const AMAP_KEY = (configJson as any).frontend?.amapKey || 'YOUR_AMAP_KEY_HERE'
const AMAP_SECURITY_CODE = (configJson as any).frontend?.amapSecurityJsCode || ''
const API_BASE_URL = (configJson as any).frontend?.backendBaseUrl || 'http://127.0.0.1:3000'

export default function ExploreMap() {
    const navigate = useNavigate()
    const mapRef = useRef<any>(null)
    const amapRef = useRef<any>(null)
    const markersRef = useRef<any[]>([])
    const currentLocationMarkerRef = useRef<any>(null)
    const areaMarkerRef = useRef<any>(null)

    const [map, setMap] = useState<any>(null)
    const [AMap, setAMap] = useState<any>(null)
    const [city, setCity] = useState('北京')
    const [searchInput, setSearchInput] = useState('')
    const [selectedTypes, setSelectedTypes] = useState<string[]>([])
    const [attractions, setAttractions] = useState<Attraction[]>([])
    const [selectedAttraction, setSelectedAttraction] = useState<Attraction | null>(null)
    const [isLoading, setIsLoading] = useState(false)
    const [isSidebarOpen, setIsSidebarOpen] = useState(true)
    const [isLocating, setIsLocating] = useState(false)
    const [suggestions, setSuggestions] = useState<Array<{ name: string; address: string; location?: { lng: number; lat: number } }>>([])
    const [showSuggestions, setShowSuggestions] = useState(false)
    const [activeSuggestIndex, setActiveSuggestIndex] = useState(0)
    const [favorites, setFavorites] = useState<Array<{ id: string; name: string; lng: number; lat: number; address: string }>>([])
    const [showFavorites, setShowFavorites] = useState(false)

    // 从后端加载收藏夹
    useEffect(() => {
        loadFavorites()
    }, [])

    const loadFavorites = async () => {
        try {
            const token = localStorage.getItem('token')
            if (!token) return

            const response = await fetch(`${API_BASE_URL}/api/favorites`, {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            })

            if (response.ok) {
                const data = await response.json()
                setFavorites(data || [])
            }
        } catch (e) {
            console.error('Failed to load favorites:', e)
        }
    }

    // 初始化地图
    useEffect(() => {
        if (AMAP_KEY === 'YOUR_AMAP_KEY_HERE') {
            console.error('请先在 config.json 中配置高德地图 API Key')
            return
        }

        if (AMAP_SECURITY_CODE) {
            (window as any)._AMapSecurityConfig = {
                securityJsCode: AMAP_SECURITY_CODE,
            }
        }

        AMapLoader.load({
            key: AMAP_KEY,
            version: '2.0',
            plugins: ['AMap.PlaceSearch', 'AMap.Geocoder', 'AMap.InfoWindow', 'AMap.AutoComplete'],
        })
            .then((AMapInstance) => {
                setAMap(AMapInstance)
                amapRef.current = AMapInstance

                const mapInstance = new AMapInstance.Map(mapRef.current, {
                    zoom: 11,
                    center: [116.397428, 39.90923],
                    viewMode: '3D',
                    pitch: 50,
                })

                setMap(mapInstance)
            })
            .catch((e) => {
                console.error('地图加载失败:', e)
            })

        return () => {
            map?.destroy()
        }
    }, [])

    const handleSearch = () => {
        if (!map || !AMap) return

        const raw = searchInput.trim()

        if (!raw) {
            setIsLoading(true)
            clearMarkers()
            searchAttractionsByCity(city, '景点')
            setShowSuggestions(false)
            return
        }

        if (suggestions.length > 0) {
            selectSuggestion(suggestions[0])
            return
        }

        fetchSuggestions(raw)
    }

    const fetchSuggestions = (keyword: string) => {
        if (!AMap) return
        const auto = new AMap.AutoComplete({ city })
        auto.search(keyword, (status: string, result: any) => {
            if (status === 'complete' && result.tips?.length) {
                const list = result.tips
                    .filter((t: any) => t.location || t.name)
                    .slice(0, 8)
                    .map((t: any) => ({
                        name: t.name,
                        address: t.district || t.address || '',
                        location: t.location ? { lng: t.location.lng, lat: t.location.lat } : undefined,
                    }))
                setSuggestions(list)
                setActiveSuggestIndex(0)
                setShowSuggestions(true)

                if (list.length > 0) {
                    selectSuggestion(list[0])
                }
            } else {
                setSuggestions([])
                setShowSuggestions(false)
                setIsLoading(true)
                clearMarkers()
                searchAttractionsByCity(city, keyword)
            }
        })
    }

    const selectSuggestion = (item: { name: string; address: string; location?: { lng: number; lat: number } }) => {
        if (!AMap || !map) return
        setShowSuggestions(false)
        setSearchInput(item.name)

        const openInfoAndSearch = (lng: number, lat: number) => {
            if (areaMarkerRef.current) {
                try { areaMarkerRef.current.setMap(null) } catch { }
                areaMarkerRef.current = null
            }
            const marker = new AMap.Marker({ position: [lng, lat], title: item.name })
            marker.setMap(map)
            areaMarkerRef.current = marker
            map.setZoomAndCenter(12, [lng, lat])

            const navUrl = getNavUrl(lng, lat, item.name)
            const detailUrl = getDetailUrl(lng, lat, item.name)
            const info = new AMap.InfoWindow({
                content: `
                  <div style="padding:10px; min-width:220px;">
                    <h3 style="margin:0 0 6px 0; font-size:16px; color:#000;">${item.name}</h3>
                    <p style="margin:0 0 8px 0; color:#333;">${item.address || ''}</p>
                    <div style="display:flex; gap:8px;">
                      <a href="${navUrl}" target="_blank" style="flex:1; padding:6px 10px; background:#667eea; color:#fff; border-radius:4px; text-decoration:none; text-align:center;">导航</a>
                      <a href="${detailUrl}" target="_blank" style="flex:1; padding:6px 10px; background:#10b981; color:#fff; border-radius:4px; text-decoration:none; text-align:center;">详情</a>
                    </div>
                  </div>
                `,
                offset: new AMap.Pixel(0, -30)
            })
            info.open(map, [lng, lat])

            searchNearbyAttractions([lng, lat])
        }

        if (item.location) {
            openInfoAndSearch(item.location.lng, item.location.lat)
        } else {
            const geocoder = new AMap.Geocoder()
            geocoder.getLocation(item.name, (status: string, result: any) => {
                if (status === 'complete' && result.geocodes?.length) {
                    const loc = result.geocodes[0].location
                    openInfoAndSearch(loc.lng, loc.lat)
                }
            })
        }
    }

    const getNavUrl = (lng: number, lat: number, name?: string) => {
        const to = `${lng},${lat},${encodeURIComponent(name || '目的地')}`
        return `https://uri.amap.com/navigation?to=${to}&mode=car&policy=1&src=travel_planner&callnative=0`
    }

    const getDetailUrl = (lng: number, lat: number, name?: string) => {
        return `https://uri.amap.com/marker?position=${lng},${lat}&name=${encodeURIComponent(name || '地点')}&src=travel_planner&callnative=0`
    }

    // 添加到收藏夹
    const addToFavorites = async (id: string, name: string, lng: number, lat: number, address: string) => {
        try {
            const token = localStorage.getItem('token')
            if (!token) {
                alert('请先登录')
                return
            }

            const favorite = { id, name, lng, lat, address }
            const response = await fetch(`${API_BASE_URL}/api/favorites`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify(favorite)
            })

            if (response.ok) {
                // 重新加载收藏夹
                await loadFavorites()
            } else if (response.status === 409) {
                alert('已经收藏过了')
            } else {
                const data = await response.json()
                alert(data.error || '收藏失败')
            }
        } catch (e) {
            console.error('Failed to add favorite:', e)
            alert('收藏失败')
        }
    }

    // 从收藏夹移除
    const removeFromFavorites = async (id: string) => {
        try {
            const token = localStorage.getItem('token')
            if (!token) return

            const response = await fetch(`${API_BASE_URL}/api/favorites/${id}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            })

            if (response.ok) {
                // 重新加载收藏夹
                await loadFavorites()
            } else {
                const data = await response.json()
                alert(data.error || '删除失败')
            }
        } catch (e) {
            console.error('Failed to remove favorite:', e)
            alert('删除失败')
        }
    }

    // 检查是否已收藏
    const isFavorited = (id: string) => {
        return favorites.some(f => f.id === id)
    }

    // 跳转到收藏的位置
    const goToFavorite = (fav: { id: string; name: string; lng: number; lat: number; address: string }) => {
        if (!map || !AMap) return

        // 关闭收藏夹面板
        setShowFavorites(false)

        // 聚焦到该位置
        map.setZoomAndCenter(15, [fav.lng, fav.lat])

        // 显示信息窗口
        const navUrl = getNavUrl(fav.lng, fav.lat, fav.name)
        const detailUrl = getDetailUrl(fav.lng, fav.lat, fav.name)
        const info = new AMap.InfoWindow({
            content: `
              <div style="padding:10px; min-width:220px;">
                <h3 style="margin:0 0 6px 0; font-size:16px; color:#000;">${fav.name}</h3>
                <p style="margin:0 0 8px 0; color:#333;">${fav.address}</p>
                <div style="display:flex; gap:8px;">
                  <a href="${navUrl}" target="_blank" style="flex:1; padding:6px 10px; background:#667eea; color:#fff; border-radius:4px; text-decoration:none; text-align:center;">导航</a>
                  <a href="${detailUrl}" target="_blank" style="flex:1; padding:6px 10px; background:#10b981; color:#fff; border-radius:4px; text-decoration:none; text-align:center;">详情</a>
                </div>
              </div>
            `,
            offset: new AMap.Pixel(0, -30)
        })
        info.open(map, [fav.lng, fav.lat])
    }

    const searchNearbyAttractions = (center: [number, number]) => {
        if (!AMap || !map) return
        clearMarkers()

        const placeSearch = new AMap.PlaceSearch({ pageSize: 20 })
        let kw = '景点'
        if (selectedTypes.length > 0) {
            const typeKeywords = selectedTypes
                .map(t => ATTRACTION_TYPES.find(at => at.value === t)?.keywords)
                .filter(Boolean)
                .join('|')
            kw = `景点 ${typeKeywords}`
        }
        placeSearch.searchNearBy(kw, center, 5000, (status: string, result: any) => {
            if (status === 'complete' && result.poiList?.pois) {
                const pois = result.poiList.pois
                const list: Attraction[] = pois.map((poi: any, idx: number) => ({
                    id: poi.id || `poi-${idx}`,
                    name: poi.name,
                    type: poi.type || '景点',
                    location: { lng: poi.location.lng, lat: poi.location.lat, address: poi.address || '' } as any,
                    rating: 3.5 + Math.random() * 1.5,
                    tags: poi.type ? poi.type.split(';') : [],
                    description: poi.address || '',
                    estimatedDuration: 2,
                    distance: poi.distance ? poi.distance / 1000 : undefined,
                }))
                setAttractions(list)
                addMarkers(list)

                try {
                    const all = [...markersRef.current]
                    if (areaMarkerRef.current) all.push(areaMarkerRef.current)
                    if ((map as any).setFitView && all.length > 0) {
                        ; (map as any).setFitView(all, false, [60, 60, 60, 360])
                    } else if (list.length > 0) {
                        map.setCenter([list[0].location.lng, list[0].location.lat])
                    }
                } catch { }
            } else {
                setAttractions([])
            }
        })
    }

    const searchAttractionsByCity = (targetCity: string, keyword: string) => {
        if (!map || !AMap) return

        const placeSearch = new AMap.PlaceSearch({
            city: targetCity,
            pageSize: 20,
            pageIndex: 1,
        })

        let searchKeyword = keyword
        if (selectedTypes.length > 0) {
            const typeKeywords = selectedTypes
                .map(t => ATTRACTION_TYPES.find(at => at.value === t)?.keywords)
                .filter(Boolean)
                .join('|')
            searchKeyword = keyword ? `${keyword} ${typeKeywords}` : typeKeywords
        }

        placeSearch.search(searchKeyword, (status: string, result: any) => {
            setIsLoading(false)

            if (status === 'complete' && result.poiList?.pois) {
                const pois = result.poiList.pois
                const attractionList: Attraction[] = pois.map((poi: any, index: number) => ({
                    id: poi.id || `poi-${index}`,
                    name: poi.name,
                    type: poi.type || '景点',
                    location: {
                        lat: poi.location.lat,
                        lng: poi.location.lng,
                        address: poi.address || poi.pname + poi.cityname + poi.adname,
                    },
                    rating: 4 + Math.random(),
                    tags: poi.type ? poi.type.split(';') : [],
                    description: poi.address || '',
                    estimatedDuration: 2,
                    distance: poi.distance ? poi.distance / 1000 : undefined,
                }))

                setAttractions(attractionList)
                addMarkers(attractionList)

                if (markersRef.current.length > 0 && typeof map.setFitView === 'function') {
                    try {
                        map.setFitView(markersRef.current, false, [60, 60, 60, 360])
                    } catch (e) {
                        const center = [attractionList[0].location.lng, attractionList[0].location.lat]
                        map.setCenter(center)
                    }
                } else if (attractionList.length > 0) {
                    const center = [attractionList[0].location.lng, attractionList[0].location.lat]
                    map.setCenter(center)
                }
            } else {
                setAttractions([])
            }
        })
    }

    const applyFilters = () => {
        handleSearch()
    }

    const addMarkers = (attractionList: Attraction[]) => {
        if (!map || !AMap) return

        attractionList.forEach((attraction) => {
            const marker = new AMap.Marker({
                position: [attraction.location.lng, attraction.location.lat],
                title: attraction.name,
                icon: new AMap.Icon({
                    size: new AMap.Size(32, 32),
                    image: 'https://webapi.amap.com/theme/v1.3/markers/n/mark_r.png',
                    imageSize: new AMap.Size(32, 32),
                }),
            })

            marker.on('click', () => {
                setSelectedAttraction(attraction)
                map.setCenter([attraction.location.lng, attraction.location.lat])

                const navUrl = getNavUrl(attraction.location.lng, attraction.location.lat, attraction.name)
                const detailUrl = getDetailUrl(attraction.location.lng, attraction.location.lat, attraction.name)
                const favorited = isFavorited(attraction.id)
                const infoWindow = new AMap.InfoWindow({
                    content: `
            <div style="padding: 10px; min-width: 220px;">
              <h3 style="margin: 0 0 8px 0; font-size: 16px; color: #000;">${attraction.name}</h3>
              <p style="margin: 4px 0; color: #333; font-size: 14px;">
                ⭐ ${attraction.rating.toFixed(1)} 分
              </p>
              <p style="margin: 4px 0 8px 0; color: #333; font-size: 14px;">
                📍 ${attraction.location.address}
              </p>
              <div style="display:flex; gap:8px;">
                <a href="${navUrl}" target="_blank" style="flex:1; padding:6px 10px; background:#667eea; color:#fff; border-radius:4px; text-decoration:none; text-align:center;">导航</a>
                <a href="${detailUrl}" target="_blank" style="flex:1; padding:6px 10px; background:#10b981; color:#fff; border-radius:4px; text-decoration:none; text-align:center;">详情</a>
                <button onclick="window.toggleFavorite_${attraction.id}()" style="padding:6px 10px; background:${favorited ? '#ef4444' : '#f59e0b'}; color:#fff; border:none; border-radius:4px; cursor:pointer;">${favorited ? '★' : '☆'}</button>
              </div>
            </div>
          `,
                    offset: new AMap.Pixel(0, -30),
                })

                    // 创建全局函数供 HTML 调用
                    ; (window as any)[`toggleFavorite_${attraction.id}`] = () => {
                        if (isFavorited(attraction.id)) {
                            removeFromFavorites(attraction.id)
                        } else {
                            addToFavorites(attraction.id, attraction.name, attraction.location.lng, attraction.location.lat, attraction.location.address)
                        }
                        infoWindow.close()
                        // 重新打开以更新星标状态
                        marker.emit('click')
                    }

                infoWindow.open(map, marker.getPosition())
            })

            marker.setMap(map)
            markersRef.current.push(marker)
        })
    }

    const clearMarkers = () => {
        markersRef.current.forEach(marker => marker.setMap(null))
        markersRef.current = []
    }

    const toggleType = (type: string) => {
        setSelectedTypes(prev =>
            prev.includes(type) ? prev.filter(t => t !== type) : [...prev, type]
        )
    }

    const locateAttraction = (attraction: Attraction) => {
        if (!map) return
        setSelectedAttraction(attraction)
        map.setZoomAndCenter(15, [attraction.location.lng, attraction.location.lat])
    }

    const locateCurrentPosition = () => {
        if (!map || !AMap) return

        clearMarkers()
        if (currentLocationMarkerRef.current) {
            try { currentLocationMarkerRef.current.setMap(null) } catch { }
            currentLocationMarkerRef.current = null
        }

        setIsLocating(true)

        AMap.plugin('AMap.Geolocation', () => {
            const geolocation = new AMap.Geolocation({
                enableHighAccuracy: true,
                timeout: 10000,
                buttonPosition: 'RB',
                zoomToAccuracy: true,
            })

            geolocation.getCurrentPosition((status: string, result: any) => {
                setIsLocating(false)

                if (status === 'complete') {
                    const { lng, lat } = result.position
                    const address = result.formattedAddress || ''
                    const cityName = result.addressComponent?.city || result.addressComponent?.province || '当前位置'

                    setCity(cityName.replace('市', ''))

                    const marker = new AMap.Marker({
                        position: [lng, lat],
                        icon: new AMap.Icon({
                            size: new AMap.Size(40, 40),
                            image: '//a.amap.com/jsapi_demos/static/demo-center/icons/poi-marker-default.png',
                            imageSize: new AMap.Size(40, 40),
                        }),
                        title: '我的位置',
                    })

                    const infoWindow = new AMap.InfoWindow({
                        content: `
                                <div style="padding: 10px;">
                                    <h3 style="margin: 0 0 8px 0; color: #000; font-weight: 600;">📍 我的位置</h3>
                                    <p style="margin: 0; color: #333;">${address}</p>
                                </div>
                            `,
                        offset: new AMap.Pixel(0, -30),
                    })

                    marker.setMap(map)
                    currentLocationMarkerRef.current = marker
                    infoWindow.open(map, [lng, lat])

                    try {
                        if (typeof map.setFitView === 'function') {
                            map.setFitView([marker], false, [80, 80, 80, 80])
                        } else {
                            map.setZoomAndCenter(15, [lng, lat])
                        }
                    } catch {
                        map.setZoomAndCenter(15, [lng, lat])
                    }

                    searchAttractionsByCity(cityName.replace('市', ''), '景点')
                } else {
                    console.error('定位失败:', result)
                    alert('定位失败，请检查浏览器定位权限设置')
                }
            })
        })
    }

    useEffect(() => {
        if (map && AMap) {
            searchAttractionsByCity(city, '景点')
        }
    }, [map, AMap])

    return (
        <div className="explore-container">
            <header className="explore-header">
                <button className="back-button" onClick={() => navigate('/')}>
                    ← 返回
                </button>
                <h1>探索景点</h1>
                <div className="search-area">
                    <input
                        type="text"
                        placeholder="搜索区域或景点 (如: 故宫 / 西湖)"
                        value={searchInput}
                        onChange={(e) => {
                            setSearchInput(e.target.value)
                            const val = e.target.value.trim()
                            if (val && AMap) {
                                const auto = new AMap.AutoComplete({ city })
                                auto.search(val, (status: string, result: any) => {
                                    if (status === 'complete' && result.tips?.length) {
                                        const list = result.tips
                                            .filter((t: any) => t.location || t.name)
                                            .slice(0, 8)
                                            .map((t: any) => ({
                                                name: t.name,
                                                address: t.district || t.address || '',
                                                location: t.location ? { lng: t.location.lng, lat: t.location.lat } : undefined,
                                            }))
                                        setSuggestions(list)
                                        setActiveSuggestIndex(0)
                                        setShowSuggestions(true)
                                    } else {
                                        setSuggestions([])
                                        setShowSuggestions(false)
                                    }
                                })
                            } else {
                                setShowSuggestions(false)
                            }
                        }}
                        onKeyDown={(e) => {
                            if (e.key === 'Enter') {
                                if (showSuggestions && suggestions.length > 0) {
                                    selectSuggestion(suggestions[activeSuggestIndex])
                                } else {
                                    handleSearch()
                                }
                            }
                            if (e.key === 'ArrowDown' && suggestions.length) {
                                e.preventDefault()
                                setActiveSuggestIndex(prev => Math.min(prev + 1, suggestions.length - 1))
                            }
                            if (e.key === 'ArrowUp' && suggestions.length) {
                                e.preventDefault()
                                setActiveSuggestIndex(prev => Math.max(prev - 1, 0))
                            }
                            if (e.key === 'Escape') setShowSuggestions(false)
                        }}
                        className="search-input-unified"
                    />
                    {showSuggestions && suggestions.length > 0 && (
                        <div className="suggestions-dropdown">
                            {suggestions.map((sug, i) => (
                                <div
                                    key={i}
                                    className={`suggestion-item ${i === activeSuggestIndex ? 'active' : ''}`}
                                    onClick={() => selectSuggestion(sug)}
                                    onMouseEnter={() => setActiveSuggestIndex(i)}
                                >
                                    <div className="suggestion-name">{sug.name}</div>
                                    <div className="suggestion-address">{sug.address}</div>
                                </div>
                            ))}
                        </div>
                    )}
                    <button onClick={handleSearch} className="search-button" disabled={isLoading}>
                        {isLoading ? '搜索中...' : '🔍 搜索'}
                    </button>
                    <button
                        className="locate-button"
                        onClick={locateCurrentPosition}
                        disabled={isLocating}
                        title="定位到我的位置"
                    >
                        {isLocating ? '⏳ 定位中' : '🎯 定位'}
                    </button>
                </div>
                <div className="current-city">
                    📍 {city}
                </div>
                <button
                    className="favorites-toggle"
                    onClick={() => setShowFavorites(!showFavorites)}
                    title="收藏夹"
                >
                    ★ 收藏 ({favorites.length})
                </button>
                <button
                    className="sidebar-toggle"
                    onClick={() => setIsSidebarOpen(!isSidebarOpen)}
                >
                    {isSidebarOpen ? '◀' : '▶'}
                </button>
            </header>

            {showFavorites && (
                <div className="favorites-panel">
                    <div className="favorites-header">
                        <h3>收藏夹</h3>
                        <button onClick={() => setShowFavorites(false)}>✕</button>
                    </div>
                    <div className="favorites-list">
                        {favorites.length === 0 ? (
                            <p className="empty-hint">暂无收藏，点击景点星标按钮添加收藏</p>
                        ) : (
                            favorites.map((fav) => (
                                <div key={fav.id} className="favorite-item" onClick={() => goToFavorite(fav)}>
                                    <div className="favorite-info">
                                        <h4>{fav.name}</h4>
                                        <p>{fav.address}</p>
                                    </div>
                                    <button
                                        className="remove-favorite"
                                        onClick={(e) => {
                                            e.stopPropagation()
                                            removeFromFavorites(fav.id)
                                        }}
                                    >
                                        ✕
                                    </button>
                                </div>
                            ))
                        )}
                    </div>
                </div>
            )}

            <div className="explore-content">
                {isSidebarOpen && (
                    <aside className="explore-sidebar">
                        <div className="filter-section">
                            <h3>景点类型</h3>
                            <div className="type-filters">
                                {ATTRACTION_TYPES.map(type => (
                                    <label key={type.value} className="type-filter">
                                        <input
                                            type="checkbox"
                                            checked={selectedTypes.includes(type.value)}
                                            onChange={() => toggleType(type.value)}
                                        />
                                        <span>{type.label}</span>
                                    </label>
                                ))}
                            </div>
                            <button onClick={applyFilters} className="apply-filter-btn">
                                应用筛选
                            </button>
                        </div>

                        <div className="attractions-list">
                            <h3>景点列表 ({attractions.length})</h3>
                            {attractions.length === 0 && !isLoading && (
                                <p className="empty-hint">暂无景点数据，请尝试搜索</p>
                            )}
                            {attractions.map((attraction) => (
                                <div
                                    key={attraction.id}
                                    className={`attraction-item ${selectedAttraction?.id === attraction.id ? 'selected' : ''}`}
                                    onClick={() => locateAttraction(attraction)}
                                >
                                    <h4>{attraction.name}</h4>
                                    <div className="attraction-meta">
                                        <span className="rating">⭐ {attraction.rating.toFixed(1)}</span>
                                        {attraction.distance && (
                                            <span className="distance">📍 {attraction.distance.toFixed(1)}km</span>
                                        )}
                                    </div>
                                    <p className="attraction-address">{attraction.location.address}</p>
                                    <div className="attraction-tags">
                                        {attraction.tags.slice(0, 3).map((tag, i) => (
                                            <span key={i} className="tag">{tag}</span>
                                        ))}
                                    </div>
                                </div>
                            ))}
                        </div>
                    </aside>
                )}

                <div className="map-container">
                    <div ref={mapRef} className="amap" style={{ width: '100%', height: '100%' }} />
                </div>
            </div>
        </div>
    )
}
