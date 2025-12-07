import { BrowserRouter, Route, Routes } from 'react-router';
import { AuthProvider } from './auth/AuthProvider';
import { useAuth } from './auth/context';
import { RequireAuth } from './components/RequireAuth';
import LoginPage from './pages/LoginPage';
import SignupPage from './pages/SignupPage';

function Home() {
  const { user, logout } = useAuth();
  return (
    <main className="container">
      <h1>Hi {user?.displayName || user?.email}</h1>
      <button className="btn" onClick={logout}>
        Log out
      </button>
    </main>
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/signup" element={<SignupPage />} />
          <Route
            path="*"
            element={
              <RequireAuth>
                <Home />
              </RequireAuth>
            }
          />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
