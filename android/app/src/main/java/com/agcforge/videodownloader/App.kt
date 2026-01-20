package com.agcforge.videodownloader

import android.content.Context
import androidx.multidex.MultiDex
import androidx.multidex.MultiDexApplication

class App: MultiDexApplication() {

    companion object {
        @Volatile
        private var mInstance: App? = null

        fun getInstance(): App {
            return mInstance ?: throw IllegalStateException("App not initialized")
        }
    }
    override fun onCreate() {
        super.onCreate()
    }

    override fun attachBaseContext(base: Context?) {
        super.attachBaseContext(base)
        MultiDex.install(this)
    }
}