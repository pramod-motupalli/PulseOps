import { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import api from "../../API/api";

const Register = () => {
    const [name, setName] = useState("");
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [confirmpassword, setConfirmPassword] = useState("");
    const [loading, setLoading] = useState(false);

    const navigate = useNavigate();

    const handleSubmit = async (e) => {
        e.preventDefault();

        if (!name.trim()) {
            alert("Please enter your name!");
            return;
        }
        if (!email.trim()) {
            alert("Please enter your email!");
            return;
        }
        if (!password) {
            alert("Please enter your password!");
            return;
        }
        if (confirmpassword !== password) {
            alert("Passwords do not match!");
            return;
        }
        try {
            setLoading(true);
            const response = await api.post("/register", {
                name,
                email,
                password,
            });

            if (response.status >= 200 && response.status < 300) {
                const token = response.data?.token || response.data?.access_token;
                if (token) {
                    localStorage.setItem("token", token);
                }
                if (response.data?.user) {
                    localStorage.setItem("user", JSON.stringify(response.data.user));
                }
                alert("Account created successfully!");
                navigate("/");
            }
        } catch (error) {
            console.error("Register error:", error);
            const message = error.response?.data?.error || error.response?.data?.message || "Registration failed. Please try again.";
            alert(message);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="bg-slate-200 max-w-[2160px] w-full min-h-screen mx-auto flex items-center justify-center p-4">
            <div className="w-full max-w-md bg-white p-8 sm:p-10 border border-gray-100 shadow-xl shadow-slate-200/60 rounded-2xl">

                <div className="mb-8 text-center">
                    <h1 className="font-bold text-[#002244] text-3xl mb-2">
                        Create Account
                    </h1>
                    <p className="text-gray-500 text-sm">
                        Please enter your details to register.
                    </p>
                </div>

                <form onSubmit={handleSubmit} className="flex flex-col gap-5">

                    <div className="flex flex-col gap-2">
                        <label htmlFor="name" className="text-sm font-semibold text-gray-700">
                            Full Name
                        </label>
                        <input
                            id="name"
                            name="name"
                            type="text"
                            value={name}
                            placeholder="John Doe"
                            required
                            onChange={(e) => setName(e.target.value)}
                            className="w-full px-4 py-3 border border-gray-200 rounded-xl bg-gray-50 text-gray-800 placeholder:text-gray-400 outline-none transition focus:border-[#002244] focus:ring-4 focus:ring-blue-900/10 focus:bg-white"
                        />
                    </div>

                    <div className="flex flex-col gap-2">
                        <label htmlFor="email" className="text-sm font-semibold text-gray-700">
                            Email Address
                        </label>
                        <input
                            id="email"
                            name="email"
                            type="email"
                            value={email}
                            placeholder="you@example.com"
                            autoComplete="email"
                            required
                            onChange={(e) => setEmail(e.target.value)}
                            className="w-full px-4 py-3 border border-gray-200 rounded-xl bg-gray-50 text-gray-800 placeholder:text-gray-400 outline-none transition focus:border-[#002244] focus:ring-4 focus:ring-blue-900/10 focus:bg-white"
                        />
                    </div>

                    <div className="flex flex-col gap-2">
                        <label htmlFor="password" className="text-sm font-semibold text-gray-700">
                            Password
                        </label>
                        <input
                            id="password"
                            name="password"
                            value={password}
                            type="password"
                            placeholder="Create a password"
                            autoComplete="new-password"
                            required
                            onChange={(e) => setPassword(e.target.value)}
                            className="w-full px-4 py-3 border border-gray-200 rounded-xl bg-gray-50 text-gray-800 placeholder:text-gray-400 outline-none transition focus:border-[#002244] focus:ring-4 focus:ring-blue-900/10 focus:bg-white"
                        />
                    </div>

                    <div className="flex flex-col gap-2">
                        <label htmlFor="confirmpassword" className="text-sm font-semibold text-gray-700">
                            Confirm Password
                        </label>
                        <input
                            id="confirmpassword"
                            name="confirmpassword"
                            value={confirmpassword}
                            type="password"
                            placeholder="Confirm your password"
                            autoComplete="new-password"
                            required
                            onChange={(e) => setConfirmPassword(e.target.value)}
                            className="w-full px-4 py-3 border border-gray-200 rounded-xl bg-gray-50 text-gray-800 placeholder:text-gray-400 outline-none transition focus:border-[#002244] focus:ring-4 focus:ring-blue-900/10 focus:bg-white"
                        />
                    </div>

                    <button
                        type="submit"
                        disabled={loading}
                        className="w-full bg-[#002244] text-white font-semibold py-3 rounded-xl shadow-md shadow-blue-900/20 hover:bg-[#003366] hover:shadow-lg active:scale-[0.98] transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        {loading ? "Creating Account..." : "Create Account"}
                    </button>
                </form>

                <p className="text-center text-sm text-gray-500 mt-8">
                    Already have an account?{" "}
                    <Link
                        to="/"
                        className="font-semibold text-[#002244] hover:text-blue-600 transition"
                    >
                        Sign in
                    </Link>
                </p>

            </div>
        </div>
    );
};

export default Register;
