import React from 'react'
import { UserData } from '@/app/contexts/AuthContext';
import { notFound } from 'next/navigation';

const API_ORIGIN = process.env.NEXT_PUBLIC_API_ORIGIN;

const UserPage = async ({ params }: { params: Promise<{ username: string }> }) => {
    const { username } = await params;
    let user: UserData;
    try {
        const res = await fetch(`${API_ORIGIN}/user/${username}`);
        if (!res.ok) return notFound();

        user = await res.json();
        if (!user) return notFound();
    }
    catch (err) {
        console.error("Failed to fetch user:", err);
        return notFound();
    }

    const pictureURL: string = user.picture_url?.replace("s96-c", "s256-c") || "https://cdn-icons-png.flaticon.com/512/149/149071.png";

    return (
        <div className="flex flex-row h-full bg-surface-tonal-0">
            <div className="flex-1 flex flex-col h-full bg-surface-0">
                <img
                    src={pictureURL}
                    className="w-full"
                    alt="User picture"
                />
                <div className="flex flex-col p-2">
                    <span className="text-5xl">{user.username}</span>
                    <div className="flex justify-between">
                        <span>Privileage: </span>
                        <span>{user.privileage}</span>
                    </div>
                    <div className="flex justify-between">
                        <span>Joined at: </span>
                        <span>{user.joined_at}</span>
                    </div>
                    <div className="flex justify-between">
                        <span>Last seen: </span>
                        <span>{user.last_seen}</span>
                    </div>
                </div>
            </div>
            <div className="flex-5">
                <h1>user page</h1>
            </div>
        </div>
    )
}

export default UserPage
