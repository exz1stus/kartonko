"use client";
import React, { ReactNode, useEffect, useRef } from 'react'
import SideBar from './SideBar/SideBar'
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from '@/components/ui/resizable'
import { useSidebar } from '@/app/SidebarProvider';
import { ImperativePanelHandle } from 'react-resizable-panels';
import Scrollbar from "@/components/template/Scrollbar";

const ResizableGrid = ({ children }: { children: ReactNode }) => {
    const { isOpen, setOpen } = useSidebar();
    const sidebarRef = useRef<ImperativePanelHandle>(null);

    useEffect(() => {
        if (sidebarRef.current) {
            if (isOpen) sidebarRef.current.expand();
            else sidebarRef.current.collapse();
        }
    }, [isOpen]);

    return (
        <ResizablePanelGroup direction="horizontal">
            <ResizablePanel
                ref={sidebarRef}
                collapsible
                minSize={5}
                defaultSize={10}
                maxSize={40}
                onCollapse={() => setOpen(false)}
                onExpand={() => setOpen(true)}
                className="bg-surface-0"
            >
                <Scrollbar>
                    <div className="h-full flex justify-center">
                        <SideBar />
                    </div>
                </Scrollbar>
            </ResizablePanel>
            <ResizableHandle className="border-r-1 border-surface-20" />
            <ResizablePanel className="h-full">
                {children}
            </ResizablePanel>
        </ResizablePanelGroup>
    )
}

export default ResizableGrid
