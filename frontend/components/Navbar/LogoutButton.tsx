import { useAuth } from "@/contexts/AuthContext";
import React from "react";

const LogoutButton: React.FC = () => {
    const { logout } = useAuth();

    return (
        <button
            className="cursor:bg-surface-10 px-2 border border-surface-20 rounded-xl"
            onClick={async () => await logout()}
        >
            Log out
        </button>
    );
};

export default LogoutButton;
