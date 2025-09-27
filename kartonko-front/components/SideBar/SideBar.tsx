"use client";
import React, { useRef } from 'react'
import Dropdown from '../template/Dropdown';
import Link from 'next/link';
import { useAuth } from '@/app/AuthContext';

const SideBar: React.FC = () => {
    const auth = useAuth();

    const authed = auth.user !== null ? (
        <Link href="/me" >Profile</Link>
    ) : null;

    const moderator = auth.user?.privileage === "Moderator" ? (
        <Dropdown text="Moderator" open={true}>
            <div className="hover:cursor-pointer">Audit Log</div>
        </Dropdown>
    ) : null;

    return (
        <aside className="h-full flex flex-col gap-5 px-4">
            <Dropdown text="" open={true}>
                <Link href="/" >Home</Link>
                {authed}
            </Dropdown>
            {moderator}
        </aside>
    )
}

export default SideBar
