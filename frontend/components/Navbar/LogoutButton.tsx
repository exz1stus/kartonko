import { useAuth } from "@/contexts/AuthContext";
import React from "react";

const LogoutButton: React.FC = () => {
    const { logout } = useAuth();

    return (
        <button
            className="bg-surface-10 px-2 border hover:border-surface-30 rounded-xl w-full cursor-pointer"
            onClick={async () => await logout()}
        >
            Log out
        </button>
    );
};

export default LogoutButton;
