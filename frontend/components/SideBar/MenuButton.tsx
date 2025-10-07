"use client";
import React from 'react'
import { Menu } from 'lucide-react'
import { useSidebar } from '@/app/contexts/SidebarContext';

const MenuButton = () => {
    const { toggle } = useSidebar();
    return (
        <>
            <div onClick={() => { toggle(); }}>
                <Menu />
            </div>
        </>
    )
}

export default MenuButton
