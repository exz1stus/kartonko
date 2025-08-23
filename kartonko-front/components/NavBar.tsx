import React from 'react'
import LoginButton from './LoginButton'

const NavBar = () => {
  return (
    <nav className="p-4 w-full flex items-center justify-between bg-surface-0">
        <span className="text-3xl font-bold">menu</span>
        <span className="text-3xl font-bold">kartonko</span>
        <LoginButton/>
    </nav>
  )
}

export default NavBar