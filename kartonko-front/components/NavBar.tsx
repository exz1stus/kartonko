import React from 'react'
import UserProfile from './UserProfile'

const NavBar = () => {
  return (
    <nav className="p-4 fixed w-full flex items-center justify-between bg-surface-0">
        <span className="text-3xl font-bold">menu</span>
        <span className="text-3xl font-bold">kartonko</span>
        <UserProfile/>
    </nav>
  )
}

export default NavBar