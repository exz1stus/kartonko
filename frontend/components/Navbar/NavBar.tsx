import { getLoggedUser } from "@/lib/user/user.server";
import MenuButton from "../SideBar/MenuButton";
import LoginButton from "./LoginButton";
import Logo from "./Logo";
import UserProfile from "./UserProfile";
import SettingsModal from "./SettingsModal";

const NavBar = async () => {
    const user = await getLoggedUser();
    const profile = !user ? <LoginButton /> : <UserProfile user={user} />;

    return (
        <nav className="interface-chrome top-0 z-10 sticky flex justify-between items-center bg-surface-0/80 p-1 border-surface-20 border-b w-full glass">
            <MenuButton />
            <Logo />
            <div className="flex items-center gap-1 p-1"><SettingsModal />{profile}</div>
        </nav>
    );
};

export default NavBar;
