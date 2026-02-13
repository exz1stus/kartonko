import React from "react";
import MenuButton from "../SideBar/MenuButton";
import LoginButton from "./LoginButton";
import Logo from "./Logo";

const NavBar = () => {
    return (
        <nav className="top-0 z-10 sticky flex justify-between items-center bg-surface-0/80 p-2 border-surface-20 border-b w-full glass">
            <MenuButton />
            <Logo />
            <LoginButton />
        </nav>
    );
};

export default NavBar;
