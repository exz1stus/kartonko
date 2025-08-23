"use client";
import React from 'react'
import { UserData, UserModal } from './UserModal'


const UserProfile: React.FC<UserData> = (user: UserData) => {
    const [modalShown, setModalShown] = React.useState(false);
    const pictureURL: string = user.picture_url || "https://cdn-icons-png.flaticon.com/512/149/149071.png";
    
    return (
        <div>
            <div className="flex gap-2">
                <div className="rounded-full" onClick={() => setModalShown(true)}>
                    <img src={pictureURL} className="h-10 w-10 rounded-full" alt="User picture" />
                </div>
            </div>
            <UserModal user={user} shown={modalShown} onClose={() => setModalShown(false)}/>
        </div>
    )
}

export default UserProfile