import { useParams } from "react-router";
export const useCourseID = () => {
    const route = useParams();
    try {
        return route.id ? BigInt(route.id) : BigInt(0);
    }
    catch (e) {
        return BigInt(0);
    }
};
