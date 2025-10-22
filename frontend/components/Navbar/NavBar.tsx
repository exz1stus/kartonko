import React from 'react'
import MenuButton from '../SideBar/MenuButton'
import LoginButton from './LoginButton'

const NavBar = () => {
    return (
        <nav
            className="top-0 z-10 sticky flex justify-between items-center bg-surface-0/80 p-2 border-surface-20 border-b w-full glass"
        >
            <MenuButton />
            <span className="left-1/2 absolute font-bold text-2xl -translate-x-1/2 transform">kartonko</span>
            <LoginButton />
        </nav>
    )
}

export default NavBar