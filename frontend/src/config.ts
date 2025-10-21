import configJson from './config.json'

interface Config {
    apiBaseUrl: string
    backendBaseUrl?: string
}

// 在构建时静态加载配置，移除运行时 fetch 依赖
const config: Config = {
    apiBaseUrl:
        (configJson as any).apiBaseUrl || (configJson as any).backendBaseUrl || 'http://127.0.0.1:3000',
    backendBaseUrl: (configJson as any).backendBaseUrl
}

export async function loadConfig(): Promise<Config> {
    return config
}

export function getConfig(): Config {
    return config
}

export function getApiUrl(path: string): string {
    return `${config.apiBaseUrl}${path}`
}
