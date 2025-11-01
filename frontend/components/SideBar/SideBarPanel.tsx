"use client";
import { useSidebar } from '@/app/contexts/SidebarContext';
import React, { useEffect, useRef } from 'react'
import { ImperativePanelHandle } from 'react-resizable-panels';
import { ResizablePanel } from '../ui/resizable';
import Scrollbar from '../template/Scrollbar';
import SideBar from './SideBar';

const SideBarPanel = () => {
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
    )
}

export default SideBarPanel
