import { NextPage } from 'next';
import MainLayout from '../../components/layout/MainLayout';

const ImportPage: NextPage = () => {
  return (
    <MainLayout>
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-2xl font-bold mb-4">Batch Import</h1>
        <p>Batch import interface will go here.</p>
      </div>
    </MainLayout>
  );
};

export default ImportPage;
