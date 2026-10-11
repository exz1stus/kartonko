"use client";
import { useTranslations } from "@/contexts/LocaleContext";

interface Props {
    time: string;
}

const TimeField = ({ time }: Props) => {
    const date = new Date(time);
    const { language } = useTranslations();

    const locale = language === "mk" ? "mk-MK" : language === "rue" ? "rue" : "en-GB";
    return <div>{date.toLocaleString(locale, { hour12: false })}</div>;
};

export default TimeField;
