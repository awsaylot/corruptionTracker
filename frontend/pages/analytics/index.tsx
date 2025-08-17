import { NextPage } from 'next';
import MainLayout from '../../components/layout/MainLayout';

const AnalyticsPage: NextPage = () => {
  return (
    <MainLayout>
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-2xl font-bold mb-4">Analytics Dashboard</h1>
        <p>Analytics overview and dashboard will go here.</p>
      </div>
    </MainLayout>
  );
};

export default AnalyticsPage;
