import React, { ReactNode } from "react";
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@/components/ui/resizable";
import SideBarPanel from "@/components/SideBar/SideBarPanel";

const ResizableGrid = ({ children }: { children: ReactNode }) => {
    return (
        <ResizablePanelGroup direction="horizontal">
            <SideBarPanel />
            <ResizableHandle className="border-surface-20 border-r-1" />
            <ResizablePanel defaultSize={95} className="h-full overflow-hidden">
                {children}
            </ResizablePanel>
        </ResizablePanelGroup>
    );
};

export default ResizableGrid;
