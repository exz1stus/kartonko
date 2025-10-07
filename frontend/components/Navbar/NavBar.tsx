import React from 'react'
import MenuButton from '../SideBar/MenuButton'
import LoginButton from './LoginButton'

const NavBar = () => {
    return (
        <nav className="border-b-1 border-surface-20 z-10 sticky top-0 p-2 w-full flex items-center justify-between bg-surface-0">
            <MenuButton />
            <span className="text-2xl font-bold">kartonko</span>
            <LoginButton />
        </nav>
    )
}

export default NavBar