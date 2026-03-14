"use client";
import { useSidebar } from "@/contexts/SidebarContext";
import React, { useEffect, useRef } from "react";
import { ImperativePanelHandle } from "react-resizable-panels";
import { ResizablePanel } from "../ui/resizable";
import Scrollbar from "../template/Scrollbar";

const SideBarPanel = ({ sidebar }: { sidebar: React.ReactNode }) => {
    const { isOpen, setOpen } = useSidebar();
    const sidebarPanRef = useRef<ImperativePanelHandle>(null);

    useEffect(() => {
        if (sidebarPanRef.current) {
            if (isOpen) sidebarPanRef.current.expand();
            else sidebarPanRef.current.collapse();
        }
    }, [isOpen]);

    return (
        <ResizablePanel
            ref={sidebarPanRef}
            collapsible
            minSize={5}
            defaultSize={7}
            maxSize={100}
            onCollapse={() => setOpen(false)}
            onExpand={() => setOpen(true)}
            className="bg-surface-0/80 backdrop-filter backdrop-blur-md"
        >
            <Scrollbar>
                <div className="flex justify-center h-full">{sidebar}</div>
            </Scrollbar>
        </ResizablePanel>
    );
};

export default SideBarPanel;
