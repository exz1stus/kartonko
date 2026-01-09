"use client";
import React, { useState } from "react";
import { UserData, UserModal } from "./UserModal";

const UserProfile: React.FC<UserData> = (user: UserData) => {
    const [modalShown, setModalShown] = useState(false);
    const pictureURL: string =
        user.picture_url || "https://cdn-icons-png.flaticon.com/512/149/149071.png";

    return (
        <div>
            <div className="flex gap-2">
                <div className="rounded-full" onClick={() => setModalShown(true)}>
                    <img src={pictureURL} className="rounded-full w-10 h-10" alt="User picture" />
                </div>
            </div>
            <UserModal user={user} shown={modalShown} onClose={() => setModalShown(false)} />
        </div>
    );
};

export default UserProfile;
