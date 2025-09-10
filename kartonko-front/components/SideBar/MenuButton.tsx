"use client";
import React from 'react'
import { Menu } from 'lucide-react'
import MenuBar from './SideMenu';

const MenuButton = () => {
    const [isOpen, setOpen] = React.useState(false);
    return (
        <>
            <div onClick={() => setOpen(true)}>
                <Menu />
            </div>
            <MenuBar isOpen={isOpen} onClose={() => setOpen(false)} />
        </>
    )
}

export default MenuButton
