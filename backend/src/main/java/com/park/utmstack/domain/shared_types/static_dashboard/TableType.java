package com.park.utmstack.domain.shared_types.static_dashboard;

import java.util.ArrayList;
import java.util.List;

public class TableType {
    private List<String> columns = new ArrayList<>();
    private List<List<String>> rows = new ArrayList<>();

    public List<String> getColumns() {
        return columns;
    }

    public void setColumns(List<String> columns) {
        this.columns = columns;
    }

    public List<List<String>> getRows() {
        return rows;
    }

    public void setRows(List<List<String>> rows) {
        this.rows = rows;
    }

    public TableType addColumn(String column) {
        columns.add(column);
        return this;
    }

    public TableType addRow(List<String> cells) {
        rows.add(cells);
        return this;
    }
}
