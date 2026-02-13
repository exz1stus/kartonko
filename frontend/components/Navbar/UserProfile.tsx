"use client";
import React, { useState } from "react";
import { UserModal } from "./UserModal";
import UserPicture from "./UserPicture";
import { UserData } from "@/app/lib/user";

interface Props {
    user: UserData;
}

const UserProfile = ({ user }: Props) => {
    const [modalShown, setModalShown] = useState(false);

    return (
        <div>
            <div className="m-2" onClick={() => setModalShown(true)}>
                <UserPicture user={user} />
            </div>
            <UserModal user={user} shown={modalShown} onClose={() => setModalShown(false)} />
        </div>
    );
};

export default UserProfile;
