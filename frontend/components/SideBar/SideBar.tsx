"use client";
import Dropdown from '../template/Dropdown';
import Link from 'next/link';
import { useAuth } from '@/app/contexts/AuthContext';
import { useHover } from '@/app/contexts/HoverContex';

const SideBar: React.FC = () => {
    const auth = useAuth();
    const { ref } = useHover();

    const authed = auth.user !== null ? (
        <Link href="/me" >Profile</Link>
    ) : null;

    const moderator = auth.user?.privileage === "Moderator" ? (
        <Dropdown text="Moderator" open={true}>
            <Link href="/log" >Log</Link>
        </Dropdown>
    ) : null;

    return (
        <aside ref={ref} className="flex flex-col items-center gap-5 px-4 w-full h-full">
            <Dropdown text="" open={true}>
                <Link href="/" >Home</Link>
                {authed}
            </Dropdown>
            {moderator}
        </aside>
    )
}

export default SideBar
