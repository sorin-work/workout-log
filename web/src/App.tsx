import { BrowserRouter, Route, Routes } from 'react-router';
import { AuthProvider } from './auth/AuthProvider';
import { Layout } from './components/Layout';
import { RequireAuth } from './components/RequireAuth';
import ExercisesPage from './pages/ExercisesPage';
import HomePage from './pages/HomePage';
import LoginPage from './pages/LoginPage';
import NotFoundPage from './pages/NotFoundPage';
import ProgramsPage from './pages/ProgramsPage';
import SettingsPage from './pages/SettingsPage';
import SignupPage from './pages/SignupPage';
import TemplateEditPage from './pages/TemplateEditPage';
import TemplatesPage from './pages/TemplatesPage';
import WorkoutPage from './pages/WorkoutPage';
import WorkoutsPage from './pages/WorkoutsPage';

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/signup" element={<SignupPage />} />
          <Route
            element={
              <RequireAuth>
                <Layout />
              </RequireAuth>
            }
          >
            <Route index element={<HomePage />} />
            <Route path="exercises" element={<ExercisesPage />} />
            <Route path="programs" element={<ProgramsPage />} />
            <Route path="settings" element={<SettingsPage />} />
            <Route path="templates" element={<TemplatesPage />} />
            <Route path="templates/new" element={<TemplateEditPage />} />
            <Route path="templates/:id/edit" element={<TemplateEditPage />} />
            <Route path="workouts" element={<WorkoutsPage />} />
            <Route path="workouts/:id" element={<WorkoutPage />} />
            <Route path="*" element={<NotFoundPage />} />
          </Route>
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
