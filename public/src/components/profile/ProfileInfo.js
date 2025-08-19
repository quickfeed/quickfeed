import React from 'react';
import { useAppState } from '../../overmind';
const ProfileInfo = ({ setEditing }) => {
    const self = useAppState().self;
    return (React.createElement(React.Fragment, null,
        React.createElement("div", { className: 'card-text text-center' },
            React.createElement("h2", { className: 'mb-4' }, self.Name)),
        React.createElement("div", { className: 'card-text text-center' },
            React.createElement("i", { className: 'fa fa-envelope text-muted' }),
            React.createElement("span", { className: 'ml-3' }, self.Email)),
        React.createElement("div", { className: 'card-text text-center' },
            React.createElement("i", { className: 'fa fa-graduation-cap text-muted' }),
            React.createElement("span", { className: 'ml-3' }, self.StudentID)),
        React.createElement("span", { className: "badge float-right clickable", onClick: () => setEditing(true) },
            React.createElement("i", { className: 'fa fa-edit' }))));
};
export default ProfileInfo;
