"use client";
import { useState } from "react";
import { UserModal } from "./UserModal";
import UserPicture from "./UserPicture";
import { UserDataResponse } from "@/lib/api/generated/model";
interface Props {
    user: UserDataResponse;
}

const UserProfile = ({ user }: Props) => {
    const [modalShown, setModalShown] = useState(false);

    return (
        <div>
            <div onClick={() => setModalShown(true)}>
                <UserPicture user={user} />
            </div>
            <UserModal
                user={user}
                shown={modalShown}
                onClose={() => setModalShown(false)}
            />
        </div>
    );
};

export default UserProfile;
