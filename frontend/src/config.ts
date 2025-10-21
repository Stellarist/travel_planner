interface Config {
    apiBaseUrl: string
}

let config: Config | null = null

export async function loadConfig(): Promise<Config> {
    if (config) {
        return config
    }

    try {
        const response = await fetch('/config.json')
        config = await response.json()
        return config as Config
    } catch (error) {
        console.error('Failed to load config, using defaults:', error)
        // 默认配置
        config = {
            apiBaseUrl: 'http://127.0.0.1:3000'
        }
        return config
    }
}

export function getConfig(): Config {
    if (!config) {
        throw new Error('Config not loaded. Call loadConfig() first.')
    }
    return config
}

export function getApiUrl(path: string): string {
    const cfg = getConfig()
    return `${cfg.apiBaseUrl}${path}`
}
