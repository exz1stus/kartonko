"use client";
import React from "react";
import { Menu } from "lucide-react";
import { useSidebar } from "@/contexts/SidebarContext";

const MenuButton = () => {
    const { toggle } = useSidebar();
    return (
        <>
            <div
                onClick={() => {
                    toggle();
                }}
                className="hover:bg-surface-10 p-2 rounded-full cursor-pointer"
            >
                <Menu />
            </div>
        </>
    );
};

export default MenuButton;
