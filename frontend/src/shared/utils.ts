import configJson from '../../../config.json'
import type { Config } from './types'
import { useEffect, useRef, useState } from 'react'

export const config: Config = {
  apiBaseUrl:
    (configJson as any).frontend?.apiBaseUrl || (configJson as any).apiBaseUrl || (configJson as any).backendBaseUrl || '/api',
  backendBaseUrl: (configJson as any).frontend?.backendBaseUrl || (configJson as any).backendBaseUrl || '/api'
}

export function getApiUrl(path: string): string {
  // 如果 path 已经以 /api 开头，直接返回
  if (path.startsWith('/api/')) {
    return path
  }
  // 如果 apiBaseUrl 是相对路径，确保正确拼接
  const baseUrl = config.apiBaseUrl.endsWith('/') ? config.apiBaseUrl.slice(0, -1) : config.apiBaseUrl
  return `${baseUrl}${path}`
}

export const formatAmount = (amount: number): string => {
  return amount >= 10000 ? (amount / 10000).toFixed(1) + '万' : amount.toFixed(2)
}

export const formatDate = (dateString: string): string => {
  const d = new Date(dateString)
  return `${d.getFullYear()}-${d.getMonth() + 1}-${d.getDate()}`
}

export const getDefaultDateRange = (days: number = 3) => {
  const today = new Date()
  const later = new Date(today)
  later.setDate(today.getDate() + days)
  return {
    start: today.toISOString().split('T')[0],
    end: later.toISOString().split('T')[0]
  }
}

export const getTodayDate = () => new Date().toISOString().slice(0, 10)

export const getDateDaysAgo = (days: number) => {
  const date = new Date()
  date.setDate(date.getDate() - days)
  return date.toISOString().slice(0, 10)
}

export const buildUrl = (path: string, params?: Record<string, string | undefined>): string => {
  if (!params) return getApiUrl(path)
  const q = new URLSearchParams()
  Object.entries(params).forEach(([k, v]) => {
    if (v !== undefined && v !== '') q.set(k, v)
  })
  const qs = q.toString()
  return getApiUrl(path) + (qs ? `?${qs}` : '')
}

const authHeaders = (): HeadersInit => {
  const token = localStorage.getItem('token')
  return token ? { Authorization: `Bearer ${token}` } : {}
}

export const apiGet = async (path: string, params?: Record<string, string | undefined>) => {
  const res = await fetch(buildUrl(path, params), { headers: authHeaders() })
  return res.json()
}

export const apiPost = async (path: string, body: any) => {
  const res = await fetch(getApiUrl(path), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify(body),
  })
  return res.json()
}

type SpeechRecognitionOptions = {
  lang?: string
  onInterim?: (text: string) => void
  onFinal?: (text: string) => void
}

export function useSpeechRecognition(opts: SpeechRecognitionOptions = {}) {
  const { lang = 'zh-CN', onInterim, onFinal } = opts
  const recognitionRef = useRef<any>(null)
  const [isListening, setIsListening] = useState(false)
  const [recognizedText, setRecognizedText] = useState('')

  const start = () => {
    if (!('webkitSpeechRecognition' in window) && !('SpeechRecognition' in window)) {
      return
    }
    const SpeechRecognition = (window as any).webkitSpeechRecognition || (window as any).SpeechRecognition
    try {
      const recognition = new SpeechRecognition()
      recognition.lang = lang
      recognition.continuous = false
      recognition.interimResults = true
      recognition.maxAlternatives = 3

      recognition.onstart = () => {
        setIsListening(true)
        setRecognizedText('')
      }

      recognition.onresult = (event: any) => {
        let interim = ''
        let final = ''
        for (let i = event.resultIndex; i < event.results.length; i++) {
          const res = event.results[i]
          if (res.isFinal) final += res[0].transcript
          else interim += res[0].transcript
        }
        if (interim) {
          setRecognizedText(interim)
          onInterim?.(interim)
        }
        if (final) {
          setRecognizedText(final)
          onFinal?.(final)
          try { recognition.stop() } catch (e) { }
        }
      }

      recognition.onerror = () => setIsListening(false)
      recognition.onend = () => {
        setIsListening(false)
        recognitionRef.current = null
      }

      recognitionRef.current = recognition
      recognition.start()
    } catch (e) { }
  }

  const stop = () => {
    const r = recognitionRef.current
    if (r?.stop) {
      try { r.stop() } catch (e) { }
    }
    recognitionRef.current = null
    setIsListening(false)
  }

  const toggle = () => {
    if (isListening) stop()
    else start()
  }

  useEffect(() => {
    return () => {
      if (recognitionRef.current?.stop) {
        try { recognitionRef.current.stop() } catch (e) { }
      }
      recognitionRef.current = null
    }
  }, [])

  return { isListening, recognizedText, start, stop, toggle }
}
