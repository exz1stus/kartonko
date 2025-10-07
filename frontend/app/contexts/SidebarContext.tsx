"use client";
import { createContext, useContext, useState } from "react";

type SidebarContextType = {
    isOpen: boolean;
    setOpen: (value: boolean) => void;
    toggle: () => void;
};

const SidebarContext = createContext<SidebarContextType | undefined>(undefined);

export default function SideBarProvider({ children }: { children: React.ReactNode }) {
    const [isOpen, setOpen] = useState(true);

    const toggle = () => {
        setOpen(!isOpen);
    };

    return (
        <SidebarContext.Provider value={{ isOpen, setOpen, toggle }}>
            {children}
        </SidebarContext.Provider>
    )
}

export function useSidebar() {
    const ctx = useContext(SidebarContext);
    if (!ctx) throw new Error("useSidebar must be used within SidebarProvider");
    return ctx;
}