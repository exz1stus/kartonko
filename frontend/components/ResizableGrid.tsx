"use client";
import React, { ReactNode, useEffect, useRef } from 'react'
import SideBar from './SideBar/SideBar'
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from '@/components/ui/resizable'
import { useSidebar } from '@/app/contexts/SidebarContext';
import { ImperativePanelHandle } from 'react-resizable-panels';
import Scrollbar from "@/components/template/Scrollbar";

const ResizableGrid = ({ children }: { children: ReactNode }) => {
    const { isOpen, setOpen } = useSidebar();
    const sidebarPanRef = useRef<ImperativePanelHandle>(null);

    useEffect(() => {
        if (sidebarPanRef.current) {
            if (isOpen) sidebarPanRef.current.expand();
            else sidebarPanRef.current.collapse();
        }
    }, [isOpen]);

    return (
        <ResizablePanelGroup direction="horizontal">
            <ResizablePanel
                ref={sidebarPanRef}
                collapsible
                minSize={5}
                defaultSize={5}
                maxSize={40}
                onCollapse={() => setOpen(false)}
                onExpand={() => setOpen(true)}
            >
                <Scrollbar>
                    <div className="flex justify-center h-full">
                        <SideBar />
                    </div>
                </Scrollbar>
            </ResizablePanel>
            <ResizableHandle className="border-surface-20 border-r-1" />
            <ResizablePanel defaultSize={95} className="h-full overflow-hidden">
                {children}
            </ResizablePanel>
        </ResizablePanelGroup>
    )
}

export default ResizableGrid
