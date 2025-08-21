import React from 'react'

import { UserIcon } from "lucide-react";

const API_ORIGIN = process.env.NEXT_PUBLIC_API_ORIGIN;


interface AuthRequest {
    username: string;
    password: string;
}

const UserProfile = () => {
    let loggedIn: boolean = false;
    
    const logIn = () => {
        fetch(`${API_ORIGIN}/login`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
        })
    }

    if (!loggedIn) {
        return (
            <div className="flex gap-2">
                <span>Sign up</span>
                /
                <span>Log in</span>
            </div>
        );
    }

    return (
        <div className="bg-white rounded-full p-2">
            <UserIcon className="h-8 w-8 text-gray-600" />
        </div>
    )
}

export default UserProfile