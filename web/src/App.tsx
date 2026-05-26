import { lazy } from 'react';
import { BrowserRouter, Route, Routes } from 'react-router';
import { AuthProvider } from './auth/AuthProvider';
import { Layout } from './components/Layout';
import { RequireAuth } from './components/RequireAuth';
import HomePage from './pages/HomePage';
import LoginPage from './pages/LoginPage';
import NotFoundPage from './pages/NotFoundPage';
import SignupPage from './pages/SignupPage';
import WorkoutPage from './pages/WorkoutPage';

// Pages used less often while training are loaded on demand.
const ExercisePage = lazy(() => import('./pages/ExercisePage'));
const ExercisesPage = lazy(() => import('./pages/ExercisesPage'));
const ProgramsPage = lazy(() => import('./pages/ProgramsPage'));
const RecordsPage = lazy(() => import('./pages/RecordsPage'));
const SettingsPage = lazy(() => import('./pages/SettingsPage'));
const TemplateEditPage = lazy(() => import('./pages/TemplateEditPage'));
const TemplatesPage = lazy(() => import('./pages/TemplatesPage'));
const VolumePage = lazy(() => import('./pages/VolumePage'));
const WorkoutsPage = lazy(() => import('./pages/WorkoutsPage'));

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
            <Route path="exercises/:id" element={<ExercisePage />} />
            <Route path="programs" element={<ProgramsPage />} />
            <Route path="records" element={<RecordsPage />} />
            <Route path="volume" element={<VolumePage />} />
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
