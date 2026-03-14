"use client";
import React, { ReactNode, useEffect, useState } from "react";
import {
    ResizableHandle,
    ResizablePanel,
    ResizablePanelGroup,
} from "@/components/ui/resizable";
import SideBarPanel from "@/components/SideBar/SideBarPanel";
import useOrientation from "@/hooks/useOrientation";

const ResizableGrid = ({
    children,
    sidebar,
}: {
    children: React.ReactNode;
    sidebar: React.ReactNode;
}) => {
    const orientation = useOrientation();
    const direction = orientation === "landscape" ? "horizontal" : "vertical";
    const [mounted, setMounted] = useState(false);

    useEffect(() => {
        setMounted(true);
    }, []);

    if (!mounted) return null;

    return (
        <ResizablePanelGroup direction={direction}>
            <SideBarPanel sidebar={sidebar} />
            <ResizableHandle className="border-surface-20 landscape:border-r portrait:border-b" />
            <ResizablePanel defaultSize={95} className="h-full overflow-hidden">
                {children}
            </ResizablePanel>
        </ResizablePanelGroup>
    );
};

export default ResizableGrid;
