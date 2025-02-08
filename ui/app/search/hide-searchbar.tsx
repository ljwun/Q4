'use client';

import { Search } from 'lucide-react'
import { Card } from "@/components/ui/card"

import { useSearchBar } from './searchbar-context';

export function SearchBarActivator() {
    const { activeComponent, setActiveComponent } = useSearchBar();
    
    if (activeComponent) return null;

    return (
        <Card
            className="w-full h-full flex items-center justify-center cursor-pointer bg-gradient-to-br from-primary/10 to-secondary/10 shadow-lg"
            onClick={() => setActiveComponent(true)}
        >
            <Search className="h-6 w-6" aria-label="顯示搜尋選項" />
        </Card>
    )
}
