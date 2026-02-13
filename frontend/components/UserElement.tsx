"use client";
import { getUserById, UserData } from "@/app/lib/user";
import React, { useEffect, useState } from "react";
import UserPicture from "./Navbar/UserPicture";
import Link from "next/link";

interface Props {
    id: number;
}

const UserElement = ({ id }: Props) => {
    const [user, setUser] = useState<UserData | null>(null);

    useEffect(() => {
        getUserById(id).then(setUser);
    }, [id]);

    if (user === undefined || user === null) return <div>Not found</div>;

    return (
        <Link className="flex flex-row items-center gap-2" href={`/user/${user?.username}`}>
            <UserPicture user={user} />
            <div>{user.username}</div>
        </Link>
    );
};

export default UserElement;
