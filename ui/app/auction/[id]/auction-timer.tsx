'use client'

import { useState, useEffect } from "react"

function calculateTimeLeft(endTime: number) {
    const difference = endTime - new Date().getTime()
    if (difference > 0) {
        const days = Math.floor(difference / (1000 * 60 * 60 * 24))
        const hours = Math.floor((difference / (1000 * 60 * 60)) % 24)
        const minutes = Math.floor((difference / 1000 / 60) % 60)
        const seconds = Math.floor((difference / 1000) % 60)
        return { days, hours, minutes, seconds }
    }
    return { days: 0, hours: 0, minutes: 0, seconds: 0 }
}

export function AuctionTimer({ endTime, mode }: { endTime: number, mode: 'start' | 'end' }) {
    const [timeLeft, setTimeLeft] = useState(calculateTimeLeft(endTime))

    useEffect(() => {
        const timer = setInterval(() => {
            setTimeLeft(calculateTimeLeft(endTime))
        }, 1000)

        return () => clearInterval(timer)
    }, [endTime])

    return (
        <div className="bg-secondary/20 p-4 rounded-lg text-center">
            <h3 className="text-lg font-semibold mb-2">
                {mode === 'start' ? "距離開始拍賣" : "拍賣倒計時"}
            </h3>
            <div className="grid grid-cols-4 gap-2">
                {Object.entries(timeLeft).map(([unit, value]) => (
                    <div key={unit} className="bg-background p-2 rounded-md">
                        <div className="text-2xl font-bold">{value}</div>
                        <div className="text-xs text-muted-foreground">{unit}</div>
                    </div>
                ))}
            </div>
        </div>
    )
}