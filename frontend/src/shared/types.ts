export interface User {
    id: number
    username: string
}

export interface Expense {
    id?: string
    category: string
    amount: number
    currency?: string
    note?: string
    createdAt?: string
}

export interface DashboardProps {
    user: User
    onLogout: () => void
}

export interface Activity {
    time: string
    type: string
    name: string
    location: string
    duration: string
    cost: number
    description: string
    tips: string
}

export interface DayItinerary {
    day: number
    date: string
    activities: Activity[]
    accommodation: string
    dailyCost: number
}

export interface TripPlan {
    id: string
    destination: string
    startDate: string
    endDate: string
    itinerary: DayItinerary[]
    totalCost: number
    summary: string
    createdAt: string
}

export interface LoginProps {
    onLoginSuccess: (user: any) => void
}

export interface Config {
    apiBaseUrl: string
    backendBaseUrl?: string
}
