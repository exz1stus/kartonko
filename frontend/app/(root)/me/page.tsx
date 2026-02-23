"use client";
import { useAuth } from "@/contexts/AuthContext";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

export default function MePage() {
    const { user, login, loading } = useAuth();
    const router = useRouter();
    useEffect(() => {
        if (loading) return;
        if (!user) login();
        else router.replace(`/user/${user?.username}`);
    }, [user, loading, login]);

    return <div>Loading...</div>;
}
