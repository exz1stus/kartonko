import React from 'react'
import LoginButton from './LoginButton'
import MenuButton from '../SideBar/MenuButton'

const NavBar = () => {
  return (
    <nav className="z-10 sticky top-0 p-4 w-full flex items-center justify-between bg-surface-0">
        <MenuButton/>
        <span className="text-3xl font-bold">kartonko</span>
        <LoginButton/>
    </nav>
  )
}

export default NavBar