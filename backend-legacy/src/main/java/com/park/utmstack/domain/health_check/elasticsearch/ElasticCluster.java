package com.park.utmstack.domain.health_check.elasticsearch;

import java.util.ArrayList;
import java.util.List;
import java.util.Locale;

public class ElasticCluster {

    private final ClusterResume resume = new ClusterResume();
    private final List<ElasticNode> nodes = new ArrayList<>();

    public List<ElasticNode> getNodes() {
        return nodes;
    }

    public ClusterResume getResume() {
        resume.getDiskTotal();
        resume.getDiskUsed();
        resume.getRamMax();
        resume.getRamCurrent();
        resume.getHeapMax();
        resume.getHeapCurrent();
        return resume;
    }

    public class ClusterResume {
        private Float diskTotal;
        private Float diskUsed;
        private Float ramMax;
        private Float ramCurrent;
        private Float heapMax;
        private Float heapCurrent;

        public Float getDiskTotal() {
            this.diskTotal = nodes.stream().map(ElasticNode::getDiskTotal).reduce((float) 0, Float::sum);
            return Float.parseFloat(String.format(Locale.US, "%.2f", this.diskTotal));
        }

        public Float getDiskUsed() {
            this.diskUsed = nodes.stream().map(ElasticNode::getDiskUsed).reduce((float) 0, Float::sum);
            return Float.parseFloat(String.format(Locale.US, "%.2f", this.diskUsed));
        }

        public Float getDiskAvailable() {
            return nodes.stream().map(ElasticNode::getDiskAvailable).reduce((float) 0, Float::sum);
        }

        public Float getDiskUsedPercent() {
            return Float.parseFloat(String.format(Locale.US, "%.2f", (diskUsed / diskTotal * 100)));
        }

        public Float getRamMax() {
            this.ramMax = nodes.stream().map(ElasticNode::getRamMax).reduce((float) 0, Float::sum);
            return Float.parseFloat(String.format(Locale.US, "%.2f", this.ramMax));
        }

        public Float getRamCurrent() {
            this.ramCurrent = nodes.stream().map(ElasticNode::getRamCurrent).reduce((float) 0, Float::sum);
            return Float.parseFloat(String.format(Locale.US, "%.2f", this.ramCurrent));
        }

        public Float getRamPercent() {
            return Float.parseFloat(String.format(Locale.US, "%.2f", (ramCurrent / ramMax * 100)));
        }

        public Float getCpuPercent() {
            return nodes.stream().map(ElasticNode::getCpuPercent).reduce((float) 0, Float::sum) / nodes.size();
        }

        public Float getHeapMax() {
            heapMax = nodes.stream().map(ElasticNode::getHeapMax).reduce((float) 0, Float::sum) / nodes.size();
            return Float.parseFloat(String.format(Locale.US, "%.2f", this.heapMax));
        }

        public Float getHeapCurrent() {
            heapCurrent = nodes.stream().map(ElasticNode::getHeapCurrent).reduce((float) 0, Float::sum) / nodes.size();
            return Float.parseFloat(String.format(Locale.US, "%.2f", this.heapCurrent));
        }

        public Float getHeapPercent() {
            return Float.parseFloat(String.format(Locale.US, "%.2f", (heapCurrent / heapMax * 100)));
        }
    }
}
