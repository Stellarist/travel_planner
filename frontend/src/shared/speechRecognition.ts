import { useEffect, useRef, useState } from 'react'

type Options = {
    lang?: string
    onInterim?: (text: string) => void
    onFinal?: (text: string) => void
}

export default function speechRecognition(opts: Options = {}) {
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
                    onInterim && onInterim(interim)
                }
                if (final) {
                    setRecognizedText(final)
                    onFinal && onFinal(final)
                    try {
                        recognition.stop()
                    } catch (e) {
                        // ignore
                    }
                }
            }

            recognition.onerror = () => {
                setIsListening(false)
            }

            recognition.onend = () => {
                setIsListening(false)
                recognitionRef.current = null
            }

            recognitionRef.current = recognition
            recognition.start()
        } catch (e) {
            // ignore startup errors
        }
    }

    const stop = () => {
        const r = recognitionRef.current
        if (r && typeof r.stop === 'function') {
            try {
                r.stop()
            } catch (e) {
                // ignore
            }
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
            if (recognitionRef.current && typeof recognitionRef.current.stop === 'function') {
                try {
                    recognitionRef.current.stop()
                } catch (e) {
                    // ignore
                }
            }
            recognitionRef.current = null
        }
    }, [])

    return { isListening, recognizedText, start, stop, toggle }
}
