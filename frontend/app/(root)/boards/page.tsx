"use client";
import { getBoard } from "@/lib/api/generated/client";
import { BoardResponse } from "@/lib/api/generated/model";
import useInfiniteScroll from "@/hooks/useInfiniteScroll";

interface Props {
    filename: string;
}

interface SearchBoardQuery {
    name?: string;
}

const BoardsPage = () => {
    const fetchEntries = async (
        query: SearchBoardQuery,
        cursor: number,
        limit: number,
    ): Promise<BoardResponse[]> => {
        return getBoard({ cursor, limit });
    };

    const { items, loading, reachedEnd, sentinelRef } = useInfiniteScroll<
        SearchBoardQuery,
        BoardResponse
    >({
        fetchFn: fetchEntries,
        query: {},
        isQueryEmpty: () => true,
    });

    const entries = items.map((entry, index) => {
        return (
            <div key={index} className="w-full lg:max-w-[50%] h-full">
                {/* <Board> */}
            </div>
        );
    });

    return (
        <div className="flex justify-center items-center w-full h-full">
            {entries}
        </div>
    );
};

export default BoardsPage;
