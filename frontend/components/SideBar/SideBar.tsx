import { getLoggedUser } from "@/lib/user/user.server";
import SidebarLinks from "@/components/SideBar/SidebarLinks";

const SideBar: React.FC = async () => {
    const user = await getLoggedUser();

    return (
        <aside className="flex portrait:flex-row landscape:flex-col portrait:flex-wrap portrait:justify-center items-center portrait:gap-2 portrait:px-2 landscape:py-2 w-full h-full">
            <SidebarLinks username={user?.username} authenticated={user !== null} moderator={user?.privilege === "Moderator"} />
        </aside>
    );
};

export default SideBar;
