"use client";
import Dropdown from "@/components/template/Dropdown";
import { useAuth } from "@/contexts/AuthContext";
import Link from "next/link";

const SideBar: React.FC = () => {
    const { user } = useAuth();

    const authed =
        user !== null ? (
            <>
                <Link href="/me">Profile</Link>
                <Link href="/upload">Upload</Link>
            </>
        ) : null;

    const moderator =
        user?.privileage === "Moderator" ? (
            <Dropdown open={true}>
                <Link href="/log">Log</Link>
            </Dropdown>
        ) : null;

    return (
        <aside className="flex portrait:flex-row landscape:flex-col portrait:flex-wrap portrait:justify-center items-center portrait:gap-2 portrait:px-2 landscape:py-2 w-full h-full">
            <Link href="/">Gallery</Link>
            {authed}
            {moderator}
        </aside>
    );
};

export default SideBar;
