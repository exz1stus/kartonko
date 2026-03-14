import React, { useRef } from "react";
import LogoutButton from "./LogoutButton";
import { UserData } from "@/lib/user";
import UsernameButton from "./UsernameButton";
import { useClickOutside } from "@/hooks/useClickOutside";
import Image from "next/image";

interface UserModalProps {
    user: UserData;
    shown: boolean;
    onClose: () => void;
}

const UserModal: React.FC<UserModalProps> = ({
    user,
    shown,
    onClose,
}: UserModalProps) => {
    const modalRef = useRef<HTMLDivElement>(null);

    useClickOutside(modalRef, onClose);

    if (!shown || !user) {
        return null;
    }

    return (
        <div
            ref={modalRef}
            className="top-0 right-0 fixed bg-surface-10 border rounded-bl-xl"
        >
            <div className="flex flex-col justify-center items-center gap-2 p-2 px-8">
                <div className="rounded-full">
                    <Image
                        src={user.picture_url}
                        className="rounded-full w-10 h-10"
                        alt="User picture"
                        width={256}
                        height={256}
                    />
                </div>
                <UsernameButton username={user.username} />
                <LogoutButton />
            </div>
        </div>
    );
};
export { UserModal, type UserData };
