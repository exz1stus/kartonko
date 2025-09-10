interface PostImageErrorSpanProps {
    error: string | null
}

const PostImageErrorSpan = ({ error }: PostImageErrorSpanProps) => {
    if (error === null || error === "") return null;

    return (
        <div className="fixed top-10 rounded-2xl bg-red-500 px-3 text-white text-center z-50">
            <span className="text-xl font-bold">{error}</span>
        </div>
    );
}

export default PostImageErrorSpan
