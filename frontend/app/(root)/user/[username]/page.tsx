import React from "react";
import { UserData } from "@/app/contexts/AuthContext";
import { notFound } from "next/navigation";
import Image from "next/image";
import GalleryServer from "@/components/Gallery/GalleryServer";

const API_ORIGIN = process.env.NEXT_PUBLIC_API_LOCAL;

const UserPage = async ({ params }: { params: { username: string } }) => {
    const { username } = params;
    let user: UserData;
    try {
        const res = await fetch(`${API_ORIGIN}/user/${username}`);
        if (!res.ok) return notFound();

        user = await res.json();
        if (!user) return notFound();
    } catch (err) {
        console.error("Failed to fetch user:", err);
        return notFound();
    }

    const pictureURL: string =
        user.picture_url?.replace("s96-c", "s256-c") ||
        "https://cdn-icons-png.flaticon.com/512/149/149071.png";

    return (
        <div className="flex flex-row h-full">
            <div className="flex flex-col flex-1 bg-surface-0/80 h-full glass">
                <Image
                    src={pictureURL}
                    className="w-full"
                    alt="User picture"
                    width={256}
                    height={256}
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
                <GalleryServer initialFetchSize={50} />
            </div>
        </div>
    );
};

export default UserPage;
