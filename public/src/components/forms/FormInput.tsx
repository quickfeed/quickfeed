import React from "react"

type FormProps = {
    prepend: string,
    name: string,
    placeholder?: string,
    defaultValue?: string | undefined,
    onChange?: (e: React.FormEvent<HTMLInputElement>) => void,
    type?: string,
    children?: React.ReactNode,
}

const FormInput = ({ prepend, name, placeholder, defaultValue, onChange, type, children }: FormProps) => {
    return (
        <div className="form-control w-full">
            <label className="label">
                <span className="label-text font-semibold">{prepend}</span>
            </label>
            <input
                className="input input-bordered w-full focus:input-primary"
                name={name}
                type={type ?? "text"}
                placeholder={placeholder}
                defaultValue={defaultValue}
                onChange={onChange}
            />
            {children}
        </div>
    )
}

export default FormInput
