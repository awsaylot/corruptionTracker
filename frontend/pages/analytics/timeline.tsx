import { NextPage } from 'next';
import MainLayout from '../../components/layout/MainLayout';

const TimelineAnalyticsPage: NextPage = () => {
  return (
    <MainLayout>
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-2xl font-bold mb-4">Timeline Analysis</h1>
        <p>Temporal analysis and timeline view will go here.</p>
      </div>
    </MainLayout>
  );
};

export default TimelineAnalyticsPage;
