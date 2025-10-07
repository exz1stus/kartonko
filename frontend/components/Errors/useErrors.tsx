//TODO : errors
const useErrors = () => {
    const showError = (error: any) => {
        console.log("Error: ", error);
    }

    return { showError }
}

export default useErrors