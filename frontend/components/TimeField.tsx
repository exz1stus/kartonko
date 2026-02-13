import React from "react";

interface Props {
    time: string;
}

const TimeField = ({ time }: Props) => {
    const date = new Date(time);

    return <div> {date.toLocaleString()}</div>;
};

export default TimeField;
