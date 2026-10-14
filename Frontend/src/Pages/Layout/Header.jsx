import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Search, ChevronDown, LogOut } from "lucide-react";

const Header = ({ title = "Dashboard", user = null }) => {
    const navigate = useNavigate();
    const [dropdownOpen, setDropdownOpen] = useState(false);
    const handleLogout = () => {
        localStorage.removeItem("token");
        localStorage.removeItem("user");
        navigate("/");
    };

    let currentUser = user;
    if (!currentUser) {
        try {
            const raw = localStorage.getItem("user");
            if (raw) {
                currentUser = JSON.parse(raw);
            }
        } catch (e) {}
    }

    const userName = currentUser?.name || currentUser?.email?.split("@")[0] || "Operator";
    const userRole = currentUser?.role || "Administrator";
    const userInitial = userName.charAt(0).toUpperCase();

    return (
        <header className="h-16 w-full bg-white border-b border-gray-200 px-6 flex items-center justify-between sticky top-0 z-30 shadow-sm">
            <div className="flex items-center gap-3">
                <h2 className="text-xl font-bold text-gray-800 tracking-tight">
                    {title}
                </h2>
            </div>

            <div className="flex items-center gap-4">
                <div className="relative">
                    <button
                        onClick={() => setDropdownOpen((prev) => !prev)}
                        className="flex items-center gap-3 p-1.5 rounded-xl hover:bg-gray-50 transition"
                    >
                        <div className="w-9 h-9 rounded-full bg-[#002244] text-white font-semibold flex items-center justify-center text-sm shadow-sm">
                            {userInitial}
                        </div>
                        <div className="hidden sm:block text-left">
                            <p className="text-sm font-semibold text-gray-800 leading-tight">
                                {userName}
                            </p>
                            <p className="text-xs text-gray-400 font-medium">
                                {userRole}
                            </p>
                        </div>
                        <ChevronDown className={`w-4 h-4 text-gray-400 transition-transform ${dropdownOpen ? "rotate-180" : ""}`} />
                    </button>

                    {dropdownOpen && (
                        <div className="absolute right-0 mt-2 w-48 bg-white border border-gray-100 rounded-xl shadow-lg py-2 z-50">
                            <div className="px-4 py-2 border-b border-gray-100 sm:hidden">
                                <p className="text-sm font-semibold text-gray-800">{userName}</p>
                                <p className="text-xs text-gray-400">{userRole}</p>
                            </div>

                            <button
                                onClick={handleLogout}
                                className="w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50 flex items-center gap-2"
                            >
                                <LogOut className="w-4 h-4 text-red-500" />
                                Sign Out
                            </button>
                        </div>
                    )}
                </div>
            </div>
        </header>
    );
};

export default Header;

