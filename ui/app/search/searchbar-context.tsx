'use client';

import { createContext, useContext, useState, ReactNode } from 'react';

type SearchBarContextType = {
  activeComponent: boolean;
  setActiveComponent: (component: boolean) => void;
};

const SearchBarContext = createContext<SearchBarContextType | undefined>(undefined);

export function SearchBarProvider({ children }: { children: ReactNode }) {
  const [activeComponent, setActiveComponent] = useState<boolean>(true);
  return (
    <SearchBarContext.Provider value={{ activeComponent: activeComponent, setActiveComponent: setActiveComponent }}>
      {children}
    </SearchBarContext.Provider>
  );
}

export function useSearchBar() {
  const context = useContext(SearchBarContext);
  if (!context) {
    throw new Error('useSearchBar must be used within a SearchBarProvider');
  }
  return context;
}
