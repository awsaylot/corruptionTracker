import { NextPage } from 'next';
import MainLayout from '../../components/layout/MainLayout';

const NetworkAnalyticsPage: NextPage = () => {
  return (
    <MainLayout>
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-2xl font-bold mb-4">Network Statistics</h1>
        <p>Network analysis and statistics will go here.</p>
      </div>
    </MainLayout>
  );
};

export default NetworkAnalyticsPage;
