import React from "react";
import MenuButton from "../SideBar/MenuButton";
import LoginButton from "./LoginButton";
import Logo from "./Logo";
import UserProfile from "./UserProfile";
import { getLoggedUserServer } from "@/lib/user";

const NavBar = async () => {
    const user = await getLoggedUserServer();
    const profile = !user ? <LoginButton /> : <UserProfile user={user} />;

    return (
        <nav className="top-0 z-10 sticky flex justify-between items-center bg-surface-0/80 p-1 border-surface-20 border-b w-full glass">
            <MenuButton />
            <Logo />
            <div className="p-1"> {profile} </div>
        </nav>
    );
};

export default NavBar;
