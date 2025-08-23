import React from 'react'
import { UserData } from '@/app/AuthContext';
import { notFound } from 'next/navigation';

const API_ORIGIN = process.env.NEXT_PUBLIC_API_ORIGIN;

const UserPage = async ({ params }: { params: Promise<{ username: string }> }) => {
    const { username } = await params;
    const res = await fetch(`${API_ORIGIN}/user/${username}`);

    if (!res.ok) {
        return notFound();
    }

    const user: UserData = await res.json();
    if (!user) return notFound();

    return (
        <div className="h-full w-[95vw] bg-surface-tonal-0">
            <img></img>
            <span>username: {user.username}</span>
        </div>
    )
}

export default UserPage
