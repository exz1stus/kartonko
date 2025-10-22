import React from 'react'
import MenuButton from '../SideBar/MenuButton'
import LoginButton from './LoginButton'

const NavBar = () => {
    return (
        <nav
            className="top-0 z-10 sticky flex justify-between items-center bg-surface-0/80 backdrop-filter backdrop-blur-md p-2 border-surface-20 border-b w-full"
        >
            <MenuButton />
            <span className="font-bold text-2xl">kartonko</span>
            <LoginButton />
        </nav>
    )
}

export default NavBar