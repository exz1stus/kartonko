"use client";
import { useClickOutside } from '@/app/clickOutside';
import React, { useRef } from 'react'
import Dropdown from '../template/Dropdown';
import Link from 'next/link';

interface MenuBarProps {
    isOpen: boolean
    onClose: () => void
}

const MenuBar: React.FC<MenuBarProps> = ({ isOpen, onClose }: MenuBarProps) => {
    const menuRef = useRef<HTMLDivElement>(null);
    useClickOutside(menuRef, onClose);
    return (
        <div ref={menuRef} className={`flex flex-col gap-5 absolute left-0 top-0 bg-surface-10 border-1 max-w-[30vw] h-screen p-4 transition-transform duration-300 ease-in-out
                ${isOpen ? 'translate-x-0' : '-translate-x-full'}`}>
            <div className="text-2xl font-bold">Menu</div>
            <Dropdown text="" open={true}>
                <Link href="/" >Home</Link>
            </Dropdown>
            <Dropdown text="Moderator" open={true}>
                <div className="hover:cursor-pointer">Audit Log</div>
            </Dropdown>
        </div>
    )
}

export default MenuBar
