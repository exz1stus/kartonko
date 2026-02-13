"use client";

import { useAuth } from "@/app/contexts/AuthContext";
import UserProfile from "./UserProfile";

const LoginButton: React.FC = () => {
    const { user, login } = useAuth();

    if (user === undefined || user === null) {
        return (
            <button onClick={async () => await login()} className="w-auto h-10 cursor-pointer">
                <span>Log in</span>
            </button>
        );
    }

    return <UserProfile user={user} />;
};

export default LoginButton;
