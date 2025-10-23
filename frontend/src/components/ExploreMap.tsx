import { useState, useEffect, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import AMapLoader from '@amap/amap-jsapi-loader'
import configJson from '../../../config.json'
import './ExploreMap.css'
import type { Attraction, TripPlan } from '../shared/types'
import { ATTRACTION_TYPES } from '../shared/constants'

const AMAP_KEY = (configJson as any).frontend?.amapKey || 'YOUR_AMAP_KEY_HERE'
const AMAP_SECURITY_CODE = (configJson as any).frontend?.amapSecurityJsCode || ''
const API_BASE_URL = (configJson as any).frontend?.backendBaseUrl || '/api'

export default function ExploreMap() {
    const navigate = useNavigate()
    const mapRef = useRef<any>(null)
    const amapRef = useRef<any>(null)
    const markersRef = useRef<any[]>([])
    const currentLocationMarkerRef = useRef<any>(null)
    const areaMarkerRef = useRef<any>(null)
    const routeOverlaysRef = useRef<any[]>([])

    const [map, setMap] = useState<any>(null)
    const [AMap, setAMap] = useState<any>(null)
    const [city, setCity] = useState('北京')
    const [searchInput, setSearchInput] = useState('')
    const [selectedTypes, setSelectedTypes] = useState<string[]>([])
    const [attractions, setAttractions] = useState<Attraction[]>([])
    const [selectedAttraction, setSelectedAttraction] = useState<Attraction | null>(null)
    const [isLoading, setIsLoading] = useState(false)
    const [isLocating, setIsLocating] = useState(false)
    const [suggestions, setSuggestions] = useState<Array<{ name: string; address: string; location?: { lng: number; lat: number } }>>([])
    const [showSuggestions, setShowSuggestions] = useState(false)
    const [activeSuggestIndex, setActiveSuggestIndex] = useState(0)
    const [favorites, setFavorites] = useState<Array<{ id: string; name: string; lng: number; lat: number; address: string }>>([])
    const [showFavorites, setShowFavorites] = useState(false)
    const [favoriteTrips, setFavoriteTrips] = useState<TripPlan[]>([])
    const [showTripRoutes, setShowTripRoutes] = useState(false)
    const [selectedTrip, setSelectedTrip] = useState<TripPlan | null>(null)

    useEffect(() => {
        loadFavorites()
        loadFavoriteTrips()
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
        }
    }

    const loadFavoriteTrips = async () => {
        try {
            const token = localStorage.getItem('token')
            if (!token) return

            const response = await fetch(`${API_BASE_URL}/api/trips/favorites/list`, {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            })

            if (response.ok) {
                const result = await response.json()

                const data = result.data || result

                if (Array.isArray(data)) {
                    setFavoriteTrips(data)
                } else {
                    setFavoriteTrips([])
                }
            } else {
                setFavoriteTrips([])
            }
        } catch (e) {
            setFavoriteTrips([])
        }
    }

    useEffect(() => {
        if (AMAP_KEY === 'YOUR_AMAP_KEY_HERE') {
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
            .catch(() => {
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

    const addToFavorites = async (id: string, name: string, lng: number, lat: number, address: string) => {
        try {
            const token = localStorage.getItem('token')
            if (!token) {
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
                await loadFavorites()
            }
        } catch (e) {
        }
    }

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
                await loadFavorites()
            }
        } catch (e) {
        }
    }

    const isFavorited = (id: string) => {
        return favorites.some(f => f.id === id)
    }

    const goToFavorite = (fav: { id: string; name: string; lng: number; lat: number; address: string }) => {
        if (!map || !AMap) return

        setShowFavorites(false)

        map.setZoomAndCenter(15, [fav.lng, fav.lat])

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

                    ; (window as any)[`toggleFavorite_${attraction.id}`] = () => {
                        if (isFavorited(attraction.id)) {
                            removeFromFavorites(attraction.id)
                        } else {
                            addToFavorites(attraction.id, attraction.name, attraction.location.lng, attraction.location.lat, attraction.location.address)
                        }
                        infoWindow.close()
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
                }
            })
        })
    }

    const clearRouteOverlays = () => {
        routeOverlaysRef.current.forEach(overlay => {
            try {
                overlay.setMap(null)
            } catch (e) {
            }
        })
        routeOverlaysRef.current = []
    }

    const drawTripRoute = async (trip: TripPlan) => {
        if (!map || !AMap) {
            return
        }

        clearRouteOverlays()
        clearMarkers()

        setSelectedTrip(trip)

        if (!trip.itinerary || trip.itinerary.length === 0) {
            return
        }

        const locations: Array<{ name: string; location: string; day: number; time: string }> = []
        trip.itinerary.forEach(day => {
            if (day.activities && day.activities.length > 0) {
                day.activities.forEach(activity => {
                    if (activity.location) {
                        const isTransportation = activity.location.includes('从') ||
                            activity.location.includes('至') ||
                            activity.location.includes('到') ||
                            activity.name.includes('地铁') ||
                            activity.name.includes('交通')

                        if (!isTransportation) {
                            locations.push({
                                name: activity.name,
                                location: activity.location,
                                day: day.day,
                                time: activity.time
                            })
                        }
                    }
                })
            }
        })

        if (locations.length === 0) {
            return
        }

        const geocoder = new AMap.Geocoder({ city: trip.destination })
        const coordinates: Array<{ lng: number; lat: number; name: string; day: number; time: string }> = []

        const colors = ['#667eea', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4', '#84cc16']
        const dayPolylines = new Map<number, any>()

        for (let i = 0; i < locations.length; i++) {
            const loc = locations[i]

            await new Promise<void>((resolve) => {
                let resolved = false
                const timeoutId = setTimeout(() => {
                    if (!resolved) {
                        resolved = true
                        resolve()
                    }
                }, 5000)

                geocoder.getLocation(loc.location, (status: string, result: any) => {
                    if (resolved) {
                        return
                    }

                    clearTimeout(timeoutId)

                    if (status === 'complete' && result.geocodes?.length) {
                        const pos = result.geocodes[0].location
                        const coord = {
                            lng: pos.lng,
                            lat: pos.lat,
                            name: loc.name,
                            day: loc.day,
                            time: loc.time
                        }
                        coordinates.push(coord)

                        const color = colors[(loc.day - 1) % colors.length]
                        const dayCoords = coordinates.filter(c => c.day === loc.day)
                        const indexInDay = dayCoords.length

                        const marker = new AMap.Marker({
                            position: [coord.lng, coord.lat],
                            title: coord.name,
                            anchor: 'bottom-center',
                            label: {
                                content: `<div style="background:${color}; color:white; padding:4px 8px; border-radius:4px; font-size:12px; font-weight:bold;">Day${loc.day}-${indexInDay}</div>`,
                                offset: new AMap.Pixel(0, -35)
                            },
                            icon: new AMap.Icon({
                                size: new AMap.Size(32, 32),
                                image: 'https://webapi.amap.com/theme/v1.3/markers/n/mark_b.png',
                                imageSize: new AMap.Size(32, 32),
                                imageOffset: new AMap.Pixel(0, 0)
                            })
                        })

                        marker.on('click', () => {
                            const info = new AMap.InfoWindow({
                                content: `
                                    <div style="padding:10px; min-width:200px;">
                                        <h3 style="margin:0 0 6px 0; font-size:16px; color:#000;">Day ${loc.day} - ${coord.time}</h3>
                                        <p style="margin:0; color:#333; font-weight:600;">${coord.name}</p>
                                    </div>
                                `,
                                offset: new AMap.Pixel(0, -30)
                            })
                            info.open(map, [coord.lng, coord.lat])
                        })

                        marker.setMap(map)
                        routeOverlaysRef.current.push(marker)

                        if (dayCoords.length > 1) {
                            const path = dayCoords.map(c => [c.lng, c.lat])

                            if (dayPolylines.has(loc.day)) {
                                const oldLine = dayPolylines.get(loc.day)!
                                oldLine.setMap(null)
                                const idx = routeOverlaysRef.current.indexOf(oldLine)
                                if (idx > -1) routeOverlaysRef.current.splice(idx, 1)
                            }

                            const polyline = new AMap.Polyline({
                                path: path,
                                strokeColor: color,
                                strokeWeight: 4,
                                strokeOpacity: 0.8,
                                lineJoin: 'round',
                                lineCap: 'round',
                            })
                            polyline.setMap(map)
                            routeOverlaysRef.current.push(polyline)
                            dayPolylines.set(loc.day, polyline)
                        }

                        if (coordinates.length === 1) {
                            map.setZoomAndCenter(15, [coord.lng, coord.lat])
                        } else {
                            map.setFitView(null, false, [60, 60, 60, 60])
                        }
                    }

                    resolved = true
                    setTimeout(resolve, 150)
                })
            })
        }

        if (coordinates.length === 0) {
            return
        }

        map.setFitView(null, false, [60, 60, 60, 60])

        setShowTripRoutes(false)
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
                    ← 返回主页
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
                    className="trip-routes-toggle"
                    onClick={() => setShowTripRoutes(!showTripRoutes)}
                    title="查看收藏的行程路线"
                >
                    🗺️ 行程路线 ({Array.isArray(favoriteTrips) ? favoriteTrips.length : 0})
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

            {showTripRoutes && (
                <div className="favorites-panel">
                    <div className="favorites-header">
                        <h3>收藏的行程</h3>
                        <button onClick={() => setShowTripRoutes(false)}>✕</button>
                    </div>
                    <div className="favorites-list">
                        {!Array.isArray(favoriteTrips) || favoriteTrips.length === 0 ? (
                            <p className="empty-hint">暂无收藏的行程，请先在规划行程页面收藏行程</p>
                        ) : (
                            favoriteTrips.map((trip) => {
                                const totalActivities = trip.itinerary?.reduce((sum, day) => sum + (day.activities?.length || 0), 0) || 0
                                return (
                                    <div key={trip.id} className="favorite-item">
                                        <div className="favorite-info" onClick={() => drawTripRoute(trip)} style={{ cursor: 'pointer', flex: 1 }}>
                                            <h4>🗺️ {trip.destination}</h4>
                                            <p>{trip.startDate} ~ {trip.endDate}</p>
                                            <p style={{ fontSize: '12px', color: '#666', marginTop: '4px' }}>
                                                {trip.itinerary?.length || 0} 天行程 · {totalActivities} 个活动
                                            </p>
                                        </div>
                                        <button
                                            className="view-route-btn"
                                            onClick={(e) => {
                                                e.stopPropagation()
                                                drawTripRoute(trip)
                                            }}
                                            style={{
                                                padding: '4px 12px',
                                                background: '#667eea',
                                                color: '#fff',
                                                border: 'none',
                                                borderRadius: '4px',
                                                cursor: 'pointer',
                                                fontSize: '12px'
                                            }}
                                        >
                                            查看路线
                                        </button>
                                    </div>
                                )
                            })
                        )}
                    </div>
                </div>
            )}

            <div className="explore-content">
                <aside className="explore-sidebar">
                    {selectedTrip && (
                        <div className="trip-route-info" style={{
                            padding: '12px',
                            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                            color: 'white',
                            borderRadius: '8px',
                            marginBottom: '16px'
                        }}>
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                <div>
                                    <h4 style={{ margin: '0 0 4px 0', fontSize: '14px' }}>当前行程路线</h4>
                                    <p style={{ margin: 0, fontSize: '16px', fontWeight: 'bold' }}>{selectedTrip.destination}</p>
                                    <p style={{ margin: '4px 0 0 0', fontSize: '12px', opacity: 0.9 }}>
                                        {selectedTrip.startDate} ~ {selectedTrip.endDate}
                                    </p>
                                </div>
                                <button
                                    onClick={() => {
                                        clearRouteOverlays()
                                        setSelectedTrip(null)
                                    }}
                                    style={{
                                        padding: '6px 12px',
                                        background: 'rgba(255,255,255,0.2)',
                                        color: 'white',
                                        border: '1px solid rgba(255,255,255,0.3)',
                                        borderRadius: '4px',
                                        cursor: 'pointer',
                                        fontSize: '12px'
                                    }}
                                >
                                    清除路线
                                </button>
                            </div>
                        </div>
                    )}
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

                <div className="map-container">
                    <div ref={mapRef} className="amap" style={{ width: '100%', height: '100%' }} />
                </div>
            </div>
        </div>
    )
}
