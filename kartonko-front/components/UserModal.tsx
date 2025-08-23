import React, { useEffect, useRef } from 'react'
import LogoutButton from './LogoutButton';
import { UserData } from '@/app/AuthContext';

interface UserModalProps {
    user: UserData;
    shown: boolean;
    onClose: () => void;
}

const UserModal: React.FC<UserModalProps> = ({ user, shown, onClose }: UserModalProps) => {
    const modalRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (modalRef.current && !modalRef.current.contains(event.target as Node)) {
                onClose();
            }
        };

        if (shown) {
            document.addEventListener("mousedown", handleClickOutside);
        }

        return () => document.removeEventListener("mousedown", handleClickOutside);
    }, [shown, onClose]);

    if (!shown || !user) {
        return null;
    }


    return (
        <div ref={modalRef} className="fixed top-0 right-0 bg-surface-10 border-surface-30 border-1 rounded-l">
            <div className="flex flex-col gap-2 p-2 pr-10 pl-10 justify-center items-center">
                <div className="rounded-full">
                    <img src={user.picture_url} className="h-10 w-10 rounded-full" alt="User picture" />
                </div>
                <div className="text-lg font-bold">{user.username}</div>
                <LogoutButton />
            </div>
        </div>
    )
}
export { UserModal, type UserData }

