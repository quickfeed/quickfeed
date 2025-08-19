import React, { useEffect } from "react";
import { useActions } from "../overmind";
const Search = ({ placeholder, setQuery, className, children }) => {
    const actions = useActions().global;
    const handleKeyUp = (e) => {
        const value = e.currentTarget.value.toLowerCase();
        setQuery ? setQuery(value) : actions.setQuery(value);
    };
    useEffect(() => {
        return () => { actions.setQuery(""); };
    }, [actions]);
    return (React.createElement("div", { className: `input-group ${className}` },
        React.createElement("input", { type: "text", className: "form-control", placeholder: placeholder ?? "Search", onKeyUp: handleKeyUp }),
        children));
};
export default Search;
