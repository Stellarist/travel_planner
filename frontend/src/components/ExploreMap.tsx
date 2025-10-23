import { useState, useEffect, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import AMapLoader from '@amap/amap-jsapi-loader'
import configJson from '../../../config.json'
import './ExploreMap.css'
import type { Attraction } from '../shared/types'
import { ATTRACTION_TYPES } from '../shared/constants'

// 常见城市列表，用于无空格输入的前缀匹配（例如：上海美食、成都火锅）
const HOT_CITIES = ['北京', '上海', '广州', '深圳', '杭州', '成都', '重庆', '武汉', '西安', '南京', '天津', '苏州', '青岛', '厦门', '长沙', '昆明', '大连', '郑州'] as const

const AMAP_KEY = (configJson as any).frontend?.amapKey || 'YOUR_AMAP_KEY_HERE'

export default function ExploreMap() {
    const navigate = useNavigate()
    const mapRef = useRef<any>(null)
    const amapRef = useRef<any>(null)
    const markersRef = useRef<any[]>([])
    const currentLocationMarkerRef = useRef<any>(null)

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

    // 初始化地图
    useEffect(() => {
        if (AMAP_KEY === 'YOUR_AMAP_KEY_HERE') {
            console.error('请先在 config.json 中配置高德地图 API Key')
            return
        }

        AMapLoader.load({
            key: AMAP_KEY,
            version: '2.0',
            plugins: ['AMap.PlaceSearch', 'AMap.Geocoder', 'AMap.InfoWindow'],
        })
            .then((AMapInstance) => {
                setAMap(AMapInstance)
                amapRef.current = AMapInstance

                const mapInstance = new AMapInstance.Map(mapRef.current, {
                    zoom: 11,
                    center: [116.397428, 39.90923], // 北京中心
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

    // 搜索景点或导航到城市
    const handleSearch = () => {
        if (!map || !AMap) return

        const raw = searchInput.trim()

        // 空输入：搜索当前城市热门景点
        if (!raw) {
            setIsLoading(true)
            clearMarkers()
            searchAttractionsByCity(city, '景点')
            return
        }

        setIsLoading(true)
        clearMarkers()

        const geocoder = new AMap.Geocoder()

        // 1) 处理包含空格的“城市 关键词”
        const parts = raw.split(/\s+/)
        if (parts.length >= 2) {
            const candCity = parts[0]
            const keyword = parts.slice(1).join(' ')
            geocoder.getLocation(candCity, (status: string, result: any) => {
                if (status === 'complete' && result.geocodes.length > 0) {
                    const loc = result.geocodes[0].location
                    setCity(candCity)
                    map.setZoomAndCenter(11, [loc.lng, loc.lat])
                    searchAttractionsByCity(candCity, keyword || '景点')
                } else {
                    // 城市无效，退化为当前城市关键词搜索
                    searchAttractionsByCity(city, raw)
                }
            })
            return
        }

        // 2) 无空格时尝试用常见城市前缀匹配，例如“上海美食”、“成都火锅”
        const matchCity = HOT_CITIES.find(cn => raw.startsWith(cn))
        if (matchCity) {
            const keyword = raw.slice(matchCity.length).trim() || '景点'
            geocoder.getLocation(matchCity, (status: string, result: any) => {
                if (status === 'complete' && result.geocodes.length > 0) {
                    const loc = result.geocodes[0].location
                    setCity(matchCity)
                    map.setZoomAndCenter(11, [loc.lng, loc.lat])
                    searchAttractionsByCity(matchCity, keyword)
                } else {
                    searchAttractionsByCity(city, raw)
                }
            })
            return
        }

        // 3) 简短输入（<=4个字）优先按城市解析，否则作为关键词
        if (raw.length <= 4) {
            geocoder.getLocation(raw, (status: string, result: any) => {
                if (status === 'complete' && result.geocodes.length > 0) {
                    const loc = result.geocodes[0].location
                    setCity(raw)
                    map.setZoomAndCenter(11, [loc.lng, loc.lat])
                    searchAttractionsByCity(raw, '景点')
                } else {
                    searchAttractionsByCity(city, raw)
                }
            })
            return
        }

        // 4) 其余情况：按当前城市的关键词搜索
        searchAttractionsByCity(city, raw)
    }

    // 搜索指定城市的景点
    const searchAttractionsByCity = (targetCity: string, keyword: string) => {
        if (!map || !AMap) return

        const placeSearch = new AMap.PlaceSearch({
            city: targetCity,
            pageSize: 20,
            pageIndex: 1,
        })

        // 构建搜索关键词
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

                // 调整地图视野：优先根据标记自适应视野
                if (markersRef.current.length > 0 && typeof map.setFitView === 'function') {
                    try {
                        // 给右侧预留更多内边距以避免被侧边栏遮挡
                        map.setFitView(markersRef.current, false, [60, 60, 60, 360])
                    } catch (e) {
                        // 兜底：以首个点为中心
                        const center = [attractionList[0].location.lng, attractionList[0].location.lat]
                        map.setCenter(center)
                    }
                } else if (attractionList.length > 0) {
                    const center = [attractionList[0].location.lng, attractionList[0].location.lat]
                    map.setCenter(center)
                }
            } else {
                setAttractions([])
                console.log('搜索结果为空')
            }
        })
    }

    // 应用筛选
    const applyFilters = () => {
        handleSearch()
    }

    // 添加标记
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

                const infoWindow = new AMap.InfoWindow({
                    content: `
            <div style="padding: 10px; min-width: 200px;">
              <h3 style="margin: 0 0 8px 0; font-size: 16px;">${attraction.name}</h3>
              <p style="margin: 4px 0; color: #666; font-size: 14px;">
                ⭐ ${attraction.rating.toFixed(1)} 分
              </p>
              <p style="margin: 4px 0; color: #666; font-size: 14px;">
                📍 ${attraction.location.address}
              </p>
              <p style="margin: 8px 0 0 0; font-size: 12px; color: #999;">
                点击侧边栏查看详情
              </p>
            </div>
          `,
                    offset: new AMap.Pixel(0, -30),
                })

                infoWindow.open(map, marker.getPosition())
            })

            marker.setMap(map)
            markersRef.current.push(marker)
        })
    }

    // 清除所有标记
    const clearMarkers = () => {
        markersRef.current.forEach(marker => marker.setMap(null))
        markersRef.current = []
    }

    // 切换景点类型筛选
    const toggleType = (type: string) => {
        setSelectedTypes(prev =>
            prev.includes(type) ? prev.filter(t => t !== type) : [...prev, type]
        )
    }

    // 定位到某个景点
    const locateAttraction = (attraction: Attraction) => {
        if (!map) return
        setSelectedAttraction(attraction)
        map.setZoomAndCenter(15, [attraction.location.lng, attraction.location.lat])
    }

    // 定位到当前位置
    const locateCurrentPosition = () => {
        if (!map || !AMap) return

        // 先清除已有的景点标记与上一次定位标记
        clearMarkers()
        if (currentLocationMarkerRef.current) {
            try { currentLocationMarkerRef.current.setMap(null) } catch { }
            currentLocationMarkerRef.current = null
        }

        setIsLocating(true)

        // 使用高德地图的定位插件
        AMap.plugin('AMap.Geolocation', () => {
            const geolocation = new AMap.Geolocation({
                enableHighAccuracy: true, // 是否使用高精度定位
                timeout: 10000, // 超时时间
                buttonPosition: 'RB', // 定位按钮的停靠位置
                zoomToAccuracy: true, // 定位成功后是否自动调整地图视野到定位点
            })

            geolocation.getCurrentPosition((status: string, result: any) => {
                setIsLocating(false)

                if (status === 'complete') {
                    const { lng, lat } = result.position
                    const address = result.formattedAddress || ''
                    const cityName = result.addressComponent?.city || result.addressComponent?.province || '当前位置'

                    // 更新城市
                    setCity(cityName.replace('市', ''))

                    // 创建当前位置标记
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
                                    <p style="margin: 0; color: #666;">${address}</p>
                                </div>
                            `,
                        offset: new AMap.Pixel(0, -30),
                    })

                    marker.setMap(map)
                    // 记录当前定位标记以便下次清除
                    currentLocationMarkerRef.current = marker
                    infoWindow.open(map, [lng, lat])

                    // 移动地图到当前位置并自适应视野
                    try {
                        if (typeof map.setFitView === 'function') {
                            map.setFitView([marker], false, [80, 80, 80, 80])
                        } else {
                            map.setZoomAndCenter(15, [lng, lat])
                        }
                    } catch {
                        map.setZoomAndCenter(15, [lng, lat])
                    }

                    // 搜索附近景点
                    searchAttractionsByCity(cityName.replace('市', ''), '景点')
                } else {
                    console.error('定位失败:', result)
                    alert('定位失败，请检查浏览器定位权限设置')
                }
            })
        })
    }

    // 初始搜索
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
                        placeholder="输入城市或景点名... (如: 上海 / 故宫)"
                        value={searchInput}
                        onChange={(e) => setSearchInput(e.target.value)}
                        onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
                        className="search-input-unified"
                    />
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
                    className="sidebar-toggle"
                    onClick={() => setIsSidebarOpen(!isSidebarOpen)}
                >
                    {isSidebarOpen ? '◀' : '▶'}
                </button>
            </header>

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
