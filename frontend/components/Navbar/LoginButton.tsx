"use client";

import { useAuth } from "@/app/contexts/AuthContext";
import UserProfile from "./UserProfile";

const LoginButton: React.FC = () => {
    const { user, login } = useAuth();

    if (!user) {
        return (
            <button onClick={async () => await login()}>
                <span>Log in</span>
            </button>
        );
    }

    return (
        <UserProfile {...user} />
    )
}

export default LoginButton