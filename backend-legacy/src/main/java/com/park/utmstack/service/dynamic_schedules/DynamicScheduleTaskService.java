package com.park.utmstack.service.dynamic_schedules;

import org.springframework.scheduling.TaskScheduler;
import org.springframework.stereotype.Service;

import java.time.Duration;
import java.time.Instant;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.ScheduledFuture;

@Service
public class DynamicScheduleTaskService {

    private final TaskScheduler scheduler;

    private final Map<String, ScheduledFuture<?>> tasks = new HashMap<>();

    public DynamicScheduleTaskService(TaskScheduler scheduler) {
        this.scheduler = scheduler;
    }

    /**
     * Gets a ${@link Map} with all scheduled tasks
     *
     * @return A ${@link Map} with all scheduled tasks
     */
    public Map<String, ScheduledFuture<?>> getTasks() {
        return tasks;
    }

    /**
     * Get a scheduled task by his identifier
     *
     * @param id Identifier of the scheduled task
     * @return A ${@link ScheduledFuture} representing the scheduled task
     */
    public ScheduledFuture<?> getTask(String id) {
        return tasks.get(id);
    }

    /**
     * Schedule the given {@link Runnable}, invoking it at the specified execution time
     * and subsequently with the given delay between the completion of one execution
     * and the start of the next.
     *
     * @param id        Task identifier
     * @param task      The ${@link Runnable} to execute whenever the trigger fires
     * @param startTime The desired first execution time for the task
     * @param delay     The delay between the completion of one execution and the start of the next
     */
    public void addTaskToScheduler(String id, Runnable task, Instant startTime, Duration delay) {
        ScheduledFuture<?> scheduledTask = scheduler.scheduleWithFixedDelay(task, startTime, delay);
        tasks.put(id, scheduledTask);
    }

    /**
     * Cancel and remove a scheduled task from local map
     *
     * @param id Identifier of scheduled task
     */
    public void removeTaskFromScheduler(String id, boolean mayInterruptIfRunning) {
        ScheduledFuture<?> task = tasks.get(id);
        if (task == null)
            return;
        task.cancel(mayInterruptIfRunning);
        tasks.remove(id);
    }
}
