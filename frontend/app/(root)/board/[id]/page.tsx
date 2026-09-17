import { getBoardId } from "@/lib/api/generated/server";

interface Props {
    id: number;
}

const ImagePage = async ({ params }: { params: Promise<Props> }) => {
    const { id } = await params;

    let board = await getBoardId(id);

    return (
        <div className="flex justify-center items-center w-full h-full">
            <title>{board.name}</title>
            <span>{board.description}</span>
        </div>
    );
};

export default ImagePage;
